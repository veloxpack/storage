package provider

import (
	"context"
	"errors"
	"io"
	"time"
)

type StorageDriver string

const (
	Filesystem StorageDriver = "fs"
	AmazonS3   StorageDriver = "s3"
)

// Storage is the storage interface.
type Storage interface {
	Save(ctx context.Context, content io.Reader, path string) error
	Stat(ctx context.Context, path string) (*Stat, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	List(ctx context.Context, path string) ([]*Stat, error)
}

// Stat contains metadata about content stored in storage.
type Stat struct {
	ModifiedTime time.Time `json:"modified_time"`
	Size         int64     `json:"size"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	ContentType  string    `json:"content_type"`
}

type StorageConfig struct {
	Driver         string `yaml:"driver" json:"driver" default:"fs" env:"STORAGE_DRIVER"`
	OutputLocation string `yaml:"output_location" json:"output_location" default:"/data" env:"STORAGE_OUTPUT_LOCATION"`
}

// ErrNotExist is a sentinel error returned by the Open and the Stat methods.
var ErrNotExist = errors.New("file does not exist")
