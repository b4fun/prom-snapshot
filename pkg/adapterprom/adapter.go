package adapterprom

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/route"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb"
	promapiv1 "github.com/prometheus/prometheus/web/api/v1"
)

// Opts set options for prometheus adapter
type Opts struct {
	StoragePath string

	Logger log.Logger

	Reg prometheus.Registerer

	// engine config
	MaxSamples int
	QueryTimeout time.Duration
	QueryConcurrency int
	LookbackDelta time.Duration

	// api config
	CORSOrigin *regexp.Regexp
}

// DefaultOpts creates default options from storage path.
func DefaultOpts(storagePath string) *Opts {
	opts := &Opts{
		StoragePath:      storagePath,

		Reg: prometheus.DefaultRegisterer,

		// see prometheus/main.go
		MaxSamples: 50000000,
		QueryTimeout: 2 * time.Minute,
		QueryConcurrency: 20,
		LookbackDelta: 5 *time.Minute,

		CORSOrigin: nil,
	}

	var promlogConfig promlog.Config
	opts.Logger = promlog.New(&promlogConfig)

	return opts
}

func (o *Opts) newQueryEngine() *promql.Engine {
	engineOpts := promql.EngineOpts{
		Logger:             log.With(o.Logger, "component", "query engine"),
		Reg:                o.Reg,
		MaxSamples:         o.MaxSamples,
		Timeout:            o.QueryTimeout,
		LookbackDelta:      o.LookbackDelta,
		ActiveQueryTracker: promql.NewActiveQueryTracker(
			o.StoragePath, o.QueryConcurrency,
			log.With(o.Logger, "component", "activeQuerytracker"),
		),
	}

	return promql.NewEngine(engineOpts)
}

func (o *Opts) newStorage() (storage.Storage, error) {
	queryStorage, err := tsdb.Open(
		o.StoragePath,
		log.With(o.Logger, "component", "storage"),
		o.Reg,
		nil, // use default config for now
	)
	if err != nil {
		return nil, fmt.Errorf("open tsdb: %w", err)
	}

	return queryStorage, nil
}

func (o *Opts) newScrapeManager(storage storage.Storage) *scrape.Manager {
	return scrape.NewManager(
		log.With(o.Logger, "component", "scrape manager"),
		storage,
	)
}

type HTTPMux interface {
	Handle(path string, handler http.Handler)
}

// NewV1API creates and registers prometheus v1 api handler.
func (o *Opts) RegisterV1API(prefix string, mux HTTPMux) error {
	queryEngine := o.newQueryEngine()
	queryStorage, err := o.newStorage()
	if err != nil {
		return err
	}
	scrapeManager := o.newScrapeManager(queryStorage)

	h := promapiv1.NewAPI(
		queryEngine,
		queryStorage,
		func(ctx context.Context) promapiv1.TargetRetriever {
			return scrapeManager
		},
		nil, nil, nil,
		promapiv1.GlobalURLOptions{},
		func(handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
		nil,
		"",
		false,
		log.With(o.Logger, "component", "promapiv1"),
		nil,
		100, 100, 100,
		o.CORSOrigin,
		nil, nil,
	)

	router := route.New()
	h.Register(router)

	prefix = fmt.Sprintf("%s/api/v1", prefix)
	mux.Handle(prefix + "/", http.StripPrefix(prefix, router))

	return nil
}