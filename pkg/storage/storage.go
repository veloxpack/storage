package storage

import (
	"context"
	"fmt"

	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"github.com/mediaprodcast/storage/pkg/storage/fs"
	"github.com/mediaprodcast/storage/pkg/storage/gcs"
	"github.com/mediaprodcast/storage/pkg/storage/s3"
)

// NewStorage creates a new storage instance
func NewStorage(driver defs.StorageDriver, cfg *defs.StorageConfig) (defs.Storage, error) {
	switch driver {
	case defs.Filesystem:
		return fs.NewStorage(fs.Config{
			Root: cfg.Filesystem.DataPath,
		}), nil
	case defs.GoogleCloudStorage:
		return gcs.NewStorage(context.Background(), cfg.GCS.CredentialsFile, cfg.GCS.Bucket)
	case defs.AmazonS3:
		return s3.NewStorage(cfg.S3)
	default:
		return nil, fmt.Errorf("storage driver %s does not exist", driver)
	}
}
