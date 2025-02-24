package storage

import (
	"github.com/veloxpack/storage/pkg/storage/fs"
	"github.com/veloxpack/storage/pkg/storage/provider"
	"github.com/veloxpack/storage/pkg/storage/rclone"
)

// StorageOption defines a functional option for configuring the storage.
type StorageOption func(*storageConfig)

type storageConfig struct {
	driver         string
	outputLocation string
}

func (cfg *storageConfig) isFileSystem() bool {
	return cfg.driver == "" || cfg.driver == string(provider.Filesystem)
}

// WithDriver sets the driver for the storage.
func WithDriver(driver string) StorageOption {
	return func(cfg *storageConfig) {
		cfg.driver = driver
	}
}

// WithOutputLocation sets the output location for the storage.
func WithOutputLocation(outputLocation string) StorageOption {
	return func(cfg *storageConfig) {
		cfg.outputLocation = outputLocation
	}
}

// NewStorage creates a new storage instance with functional options.
func NewStorage(opts ...StorageOption) provider.Storage {
	cfg := &storageConfig{
		outputLocation: "/data",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.isFileSystem() {
		return fs.NewStorage(fs.Config{Root: cfg.outputLocation})
	}

	return rclone.NewStorage(cfg.driver, cfg.outputLocation)
}
