package storage

import (
	"errors"

	"github.com/mediaprodcast/storage/pkg/storage/fs"
	"github.com/mediaprodcast/storage/pkg/storage/provider"
	"github.com/mediaprodcast/storage/pkg/storage/rclone"
)

// StorageOption defines a functional option for configuring the storage.
type StorageOption func(*storageConfig) error

type storageConfig struct {
	driver         string
	outputLocation string
}

func (cfg *storageConfig) isFileSystem() bool {
	return cfg.driver == "" || cfg.driver == string(provider.Filesystem)
}

// WithDriver sets the driver for the storage.
func WithDriver(driver string) StorageOption {
	return func(cfg *storageConfig) error {
		cfg.driver = driver
		return nil
	}
}

// WithOutputLocation sets the output location for the storage.
func WithOutputLocation(outputLocation string) StorageOption {
	return func(cfg *storageConfig) error {
		cfg.outputLocation = outputLocation
		return nil
	}
}

// NewStorage creates a new storage instance with functional options.
func NewStorage(opts ...StorageOption) (provider.Storage, error) {
	cfg := &storageConfig{}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.outputLocation == "" {
		return nil, errors.New("output location is required")
	}

	if cfg.isFileSystem() {
		return fs.NewStorage(fs.Config{Root: cfg.outputLocation}), nil
	}

	return rclone.NewStorage(cfg.driver, cfg.outputLocation), nil
}
