package storage

import (
	"context"
	"fmt"

	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"github.com/mediaprodcast/storage/pkg/storage/fs"
	"github.com/mediaprodcast/storage/pkg/storage/gcs"
	"github.com/mediaprodcast/storage/pkg/storage/s3"
)

// NewStorage creates a new storage instance
func NewStorage(cfg *pb.StorageConfig) (defs.Storage, error) {
	switch cfg.Driver {
	case pb.StorageDriver_FS:
		return fs.NewStorage(fs.Config{
			Root: cfg.Fs.DataPath,
		}), nil
	case pb.StorageDriver_GCS:
		return gcs.NewStorage(context.Background(), cfg.Gcs.CredentialsFile, cfg.Gcs.Bucket)
	case pb.StorageDriver_S3:
		return s3.NewStorage(cfg.S3)
	default:
		return nil, fmt.Errorf("storage driver %s does not exist", cfg.Driver)
	}
}
