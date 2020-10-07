package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/b4fun/prom-snapshot/pkg/snapshotsidecar/archive"
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
		newSnapshotHandler(
			logger.WithField("component", "httpServer"),
			*flagPrometheusSnapshotsDir,
			promClient,
		),
	)

	logger.Infof("http server started at %s", *flagAddr)
	if err := http.ListenAndServe(*flagAddr, mux); err != nil {
		kingpin.Fatalf("server http: %s", err)
		return
	}
}

type snapshotResponse struct {
	Error       string `json:"error,omitempty"`
	ArtifactURL string `json:"artifactURL,omitempty`
}

func responseErr(rw http.ResponseWriter, err error) {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(http.StatusInternalServerError)

	rv := snapshotResponse{
		Error: err.Error(),
	}
	_ = json.NewEncoder(rw).Encode(rv)
}

func newSnapshotHandler(
	logger logrus.FieldLogger,
	snapshotsDir string,
	promClient promclient.Client,
) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		createSnapshot, err := promClient.CreateSnapshot(r.Context())
		if err != nil {
			logger.Errorf("create snapshot failed: %s", err)
			responseErr(rw, err)
			return
		}

		snapshotFullPath := filepath.Join(snapshotsDir, createSnapshot.SnapshotName)
		logger.Infof("created snapshot at %s", snapshotFullPath)

		var snapshotArchive bytes.Buffer
		createArchive := &archive.CreateZipArchive{
			Logger:      logger,
			SnapshotDir: snapshotFullPath,
		}
		if err := createArchive.CreateTo(&snapshotArchive); err != nil {
			logger.Errorf("create snapshot archive failed: %s", err)
			responseErr(rw, err)
			return
		}

		// TODO: upload

		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		resp := snapshotResponse{
			ArtifactURL: snapshotFullPath,
		}
		_ = json.NewEncoder(rw).Encode(resp)
	})
}
