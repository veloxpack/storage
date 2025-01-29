package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	cfg := &defs.StorageConfig{
		Driver: defs.Filesystem,
		Filesystem: defs.FileSystemConfig{
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
	assert.Equal(t, cfg.Filesystem.DataPath, decoded.Filesystem.DataPath)
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
	cfg := &defs.StorageConfig{
		Driver: defs.GoogleCloudStorage,
		GCS: defs.GCSConfig{
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
	assert.Equal(t, cfg.GCS.CredentialsFile, decoded.GCS.CredentialsFile)
	assert.Equal(t, cfg.GCS.Bucket, decoded.GCS.Bucket)
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

	// Encode the config
	encoded, err := Encode(cfg)
	assert.NoError(t, err)

	// Decode the encoded string back
	decoded, err := Decode(encoded)
	assert.NoError(t, err)

	// Check that the decoded config matches the original one
	assert.Equal(t, cfg.Driver, decoded.Driver)
	assert.Equal(t, cfg.S3.Endpoint, decoded.S3.Endpoint)
	assert.Equal(t, cfg.S3.AccessKeyID, decoded.S3.AccessKeyID)
	assert.Equal(t, cfg.S3.SecretAccessKey, decoded.S3.SecretAccessKey)
	assert.Equal(t, cfg.S3.Bucket, decoded.S3.Bucket)
	assert.Equal(t, cfg.S3.EnableSSL, decoded.S3.EnableSSL)
}
