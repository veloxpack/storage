package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	t.Run("should use non-existing storage driver", func(t *testing.T) {
		_, err := NewStorage(&defs.StorageConfig{
			Driver: "not-defined-driver",
		})
		assert.EqualError(t, err, "storage driver not-defined-driver does not exist")
	})

	t.Run("should initialize Filesystem storage", func(t *testing.T) {
		cfg := &defs.StorageConfig{
			Driver: defs.Filesystem,
			Filesystem: defs.FileSystemConfig{
				DataPath: "./tmp",
			},
		}

		_, err := NewStorage(cfg)
		assert.NoError(t, err)
	})

	t.Run("should initialize S3 storage", func(t *testing.T) {
		// Skipping as it continuously fail
		// We should NOT rely on the network anymore
		t.Skip()

		cfg := &defs.StorageConfig{
			Driver: defs.AmazonS3,
			S3: defs.S3Config{
				Endpoint:        "play.min.io",
				AccessKeyID:     "Q3AM3UQ867SPQQA43P2F",
				SecretAccessKey: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
				Bucket:          uuid.New().String(),
				EnableSSL:       true,
			},
		}

		_, err := NewStorage(cfg)
		assert.NoError(t, err)
	})
}
