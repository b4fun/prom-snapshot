package promclient

import "context"

// CreateSnapshotResponse represents the create snapshot response.
type CreateSnapshotResponse struct {
	SnapshotName string
}

// Client provides interaction endpoints with prometheus.
type Client interface {
	// CreateSnapshot requests prometheus to create a snapshot. Returns snapshot name.
	CreateSnapshot(ctx context.Context) (*CreateSnapshotResponse, error)
}
