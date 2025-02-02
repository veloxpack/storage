package storage

import (
	"testing"

	"github.com/google/uuid"
	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	t.Run("should use non-existing storage driver", func(t *testing.T) {
		_, err := NewStorage(&pb.StorageConfig{
			Driver: pb.StorageDriver_UNKNOWN,
		})
		assert.EqualError(t, err, "storage driver UNKNOWN does not exist")
	})

	t.Run("should initialize Filesystem storage", func(t *testing.T) {
		cfg := &pb.StorageConfig{
			Driver: pb.StorageDriver_FS,
			Fs: &pb.FileSystemConfig{
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

		cfg := &pb.StorageConfig{
			Driver: pb.StorageDriver_S3,
			S3: &pb.S3Config{
				Endpoint:        "play.min.io",
				AccessKeyId:     "Q3AM3UQ867SPQQA43P2F",
				SecretAccessKey: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
				Bucket:          uuid.New().String(),
				EnableSsl:       true,
			},
		}

		_, err := NewStorage(cfg)
		assert.NoError(t, err)
	})
}
