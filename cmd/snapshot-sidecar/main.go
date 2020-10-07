package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/b4fun/battery/archive"
	"github.com/b4fun/prom-snapshot/pkg/snapshotsidecar/promclient"
	"github.com/b4fun/prom-snapshot/pkg/snapshotsidecar/storage"
	"github.com/b4fun/prom-snapshot/pkg/snapshotsidecar/storage/azblob"
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
	flagAzBlobContainerURL = kingpin.Flag("storage.azblob-container-url", "azblob container url").
				Required().
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
		Logger:  logger.WithField("component", "promClient"),
	}
	promClient, err := createPromClient.Create()
	if err != nil {
		kingpin.Fatalf("create prometheus api client: %w", err)
		return
	}

	createAzBlob := &azblob.StorageOption{
		Logger:                logger.WithField("component", "storage.azBlob"),
		ServicePrincipalToken: azblob.CreateServincePrincipalTokenFromEnvironment,
		ContainerURL:          *flagAzBlobContainerURL,
	}
	azBlob, err := createAzBlob.Create()
	if err != nil {
		kingpin.Fatalf("create azblob storage: %s", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle(
		"/api/v1/admin/tsdb/snapshot",
		newSnapshotHandler(
			logger.WithField("component", "httpServer"),
			*flagPrometheusSnapshotsDir,
			promClient,
			azBlob,
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
	ArtifactURL string `json:"artifactURL,omitempty"`
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
	uploader storage.Uploader,
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

		snapshotArchive := new(bytes.Buffer)
		createArchive := &archive.CreateZipArchive{
			SourceDir: snapshotFullPath,
			// under data dir
			BasePath: "data",
		}
		if err := createArchive.CompressTo(snapshotArchive); err != nil {
			logger.Errorf("create snapshot archive failed: %s", err)
			responseErr(rw, err)
			return
		}

		artifactURL, err := uploader.UploadStream(
			r.Context(),
			snapshotArchive, createSnapshot.SnapshotName,
		)
		if err != nil {
			logger.Errorf("upload snapshot archive failed: %s", err)
			responseErr(rw, err)
			return
		}
		logger.Infof("uploaded snapshot to %s", artifactURL)

		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		resp := snapshotResponse{
			ArtifactURL: artifactURL,
		}
		_ = json.NewEncoder(rw).Encode(resp)
	})
}
