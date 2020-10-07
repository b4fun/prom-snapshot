package storage

import (
	"context"
	"io"
)

// Uploader defines the upload interface.
type Uploader interface {
	// UploadStream uploads a stream with name. Returns uploaded name.
	UploadStream(ctx context.Context, stream io.Reader, name string) (string, error)
}
