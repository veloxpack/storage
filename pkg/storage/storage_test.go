package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	t.Run("should use non-existing storage driver", func(t *testing.T) {
		_, err := NewStorage("test", &defs.StorageConfig{})
		assert.EqualError(t, err, "storage driver test does not exist")
	})

	t.Run("should initialize Filesystem storage", func(t *testing.T) {
		cfg := &defs.StorageConfig{
			Filesystem: defs.FileSystemConfig{
				DataPath: "./tmp",
			},
		}

		_, err := NewStorage(defs.Filesystem, cfg)
		assert.NoError(t, err)
	})

	t.Run("should initialize S3 storage", func(t *testing.T) {
		// Skipping as it continuously fail
		// We should NOT rely on the network anymore
		t.Skip()

		cfg := &defs.StorageConfig{
			S3: defs.S3Config{
				Hostname:        "play.min.io",
				AccessKeyID:     "Q3AM3UQ867SPQQA43P2F",
				SecretAccessKey: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
				Bucket:          uuid.New().String(),
				EnableSSL:       true,
			},
		}

		_, err := NewStorage(defs.AmazonS3, cfg)
		assert.NoError(t, err)
	})
}
