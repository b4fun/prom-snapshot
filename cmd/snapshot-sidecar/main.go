package main

import (
	"net/http"

	"github.com/b4fun/prom-snapshot/pkg/snapshotsidecar/promclient"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	flagDebug = kingpin.Flag("debug", "debug?").Default("false").Bool()
	flagAddr  = kingpin.Flag("api.addr", "snapshot sidecar listen addr").
			Default(":9092").
			String()
	flagPrometheusAPIBase = kingpin.Flag("prom.api-base", "prometheus API base url").
				Default("http://0.0.0.0:9090").
				String()
	flagPrometheusSnapshotsDir = kingpin.Flag("prom.snapshots-dir", "prometheus snapshots data dir").
					Default("/prometheus/data/snapshots").
					String()
)

func main() {
	kingpin.Parse()

	logger := logrus.New()
	if *flagDebug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	createPromClient := promclient.ClientOption{
		BaseURL: *flagPrometheusAPIBase,
		Logger:  logger.WithField("component", "promclient"),
	}
	promClient, err := createPromClient.Create()
	if err != nil {
		kingpin.Fatalf("create prometheus api client: %w", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle(
		"/api/v1/admin/tsdb/snapshot",
		newSnapshotHandler(*flagPrometheusSnapshotsDir, promClient),
	)

	if err := http.ListenAndServe(*flagAddr, mux); err != nil {
		kingpin.Fatalf("server http: %s", err)
		return
	}
}

func newSnapshotHandler(
	snapshotsDir string,
	promClient promclient.Client,
) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
	})
}
