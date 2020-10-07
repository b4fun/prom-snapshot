package azblob

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/b4fun/prom-snapshot/pkg/snapshotsidecar/storage"
	"github.com/sirupsen/logrus"
)

// StorageOption defines option for creating azure blob based storage.
type StorageOption struct {
	Logger                logrus.FieldLogger
	ServicePrincipalToken func() (*adal.ServicePrincipalToken, error)
	ContainerURL          string
}

// Create creates an azblob instance.
func (s *StorageOption) Create() (*AZBlob, error) {
	logger := s.Logger
	if logger == nil {
		logger = logrus.New()
	}

	containerURL, err := url.Parse(s.ContainerURL)
	if err != nil {
		return nil, fmt.Errorf("parse ContainerURL %s: %w", s.ContainerURL, err)
	}

	if s.ServicePrincipalToken == nil {
		return nil, fmt.Errorf("ServicePrincipalToken is required")
	}
	spt, err := s.ServicePrincipalToken()
	if err != nil {
		return nil, fmt.Errorf("read service principal token: %w", err)
	}

	cred := azblob.NewTokenCredential(
		spt.OAuthToken(),
		func(credential azblob.TokenCredential) time.Duration {
			_ = spt.RefreshWithContext(context.Background())
			credential.SetToken(spt.OAuthToken())
			return time.Until(spt.Token().Expires())
		},
	)
	pipeline := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	azBlobContainerURL := azblob.NewContainerURL(*containerURL, pipeline)

	return &AZBlob{
		logger:       logger,
		containerURL: azBlobContainerURL,
	}, nil
}

// AZBlob implements azure blob based storage.
type AZBlob struct {
	logger       logrus.FieldLogger
	containerURL azblob.ContainerURL
}

var _ storage.Uploader = (*AZBlob)(nil)

func (ab *AZBlob) UploadStream(
	ctx context.Context,
	stream io.Reader,
	name string,
) (string, error) {
	blobURL := ab.containerURL.NewBlockBlobURL(name)
	ab.logger.Debugf("blob url: %s", blobURL)

	_, err := azblob.UploadStreamToBlockBlob(
		ctx,
		stream,
		blobURL,
		azblob.UploadStreamToBlockBlobOptions{},
	)
	if err != nil {
		err = fmt.Errorf("upload %s: %w", name, err)
		ab.logger.WithError(err).Error()
		return "", err
	}

	return blobURL.String(), nil
}
