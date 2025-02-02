package credentials

import (
	"testing"

	"github.com/google/uuid"
	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	cfg := &pb.StorageConfig{
		Driver: pb.StorageDriver_FS,
		Fs: &pb.FileSystemConfig{
			DataPath: "./tmp",
		},
	}

	encoded, err := Encode(cfg)

	// Test that no error occurred during encoding
	assert.NoError(t, err)

	// Test that the encoded string is not empty
	assert.NotEmpty(t, encoded)

	// Decode the encoded string and ensure that the result is the same as the original config
	decoded, err := Decode(encoded)
	// Test that no error occurred during decoding
	assert.NoError(t, err)

	// Test that the decoded config is the same as the original config
	assert.Equal(t, cfg.Driver, decoded.Driver)
	assert.Equal(t, cfg.Fs.DataPath, decoded.Fs.DataPath)
}

func TestDecodeInvalid(t *testing.T) {
	// Test an invalid Base64 string
	invalidBase64 := "invalid_base64_string"

	decoded, err := Decode(invalidBase64)

	// Test that an error occurred during decoding
	assert.Error(t, err)
	assert.Nil(t, decoded)
}

func TestEncodeDecodeConsistency(t *testing.T) {
	cfg := &pb.StorageConfig{
		Driver: pb.StorageDriver_GCS,
		Gcs: &pb.GCSConfig{
			CredentialsFile: "./path/to/credentials.json",
			Bucket:          "bucket-name",
		},
	}

	// Encode the config
	encoded, err := Encode(cfg)
	assert.NoError(t, err)

	// Decode the encoded string back
	decoded, err := Decode(encoded)
	assert.NoError(t, err)

	// Check that the decoded config matches the original one
	assert.Equal(t, cfg.Driver, decoded.Driver)
	assert.Equal(t, cfg.Gcs.CredentialsFile, decoded.Gcs.CredentialsFile)
	assert.Equal(t, cfg.Gcs.Bucket, decoded.Gcs.Bucket)
}

func TestDecodeEmptyString(t *testing.T) {
	// Test decoding an empty Base64 string
	decoded, err := Decode("")

	// Check if decoding returns an error and nil value
	assert.Error(t, err)
	assert.Nil(t, decoded)
}

// TestEncodeDecodeConsistencyForS3 checks encoding and decoding for Amazon S3 configuration
func TestEncodeDecodeConsistencyForS3(t *testing.T) {
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

	// Encode the config
	encoded, err := Encode(cfg)
	assert.NoError(t, err)

	// Decode the encoded string back
	decoded, err := Decode(encoded)
	assert.NoError(t, err)

	// Check that the decoded config matches the original one
	assert.Equal(t, cfg.Driver, decoded.Driver)
	assert.Equal(t, cfg.S3.Endpoint, decoded.S3.Endpoint)
	assert.Equal(t, cfg.S3.AccessKeyId, decoded.S3.AccessKeyId)
	assert.Equal(t, cfg.S3.SecretAccessKey, decoded.S3.SecretAccessKey)
	assert.Equal(t, cfg.S3.Bucket, decoded.S3.Bucket)
	assert.Equal(t, cfg.S3.EnableSsl, decoded.S3.EnableSsl)
}
