package defs

import (
	"context"
	"errors"
	"io"
	"time"
)

type StorageDriver string

const (
	Filesystem         StorageDriver = "fs"
	GoogleCloudStorage StorageDriver = "gcs"
	AmazonS3           StorageDriver = "s3"
)

type FileSystemConfig struct {
	DataPath string `yaml:"data_path" json:"data_path" default:"/data" env:"FS_DATA_PATH"`
}

type GCSConfig struct {
	CredentialsFile string `yaml:"credentials_file" json:"credentials_file" default:"" env:"GCS_CREDENTIALS_FILE"`
	Bucket          string `yaml:"bucket" json:"bucket" default:"" env:"GCS_BUCKET"`
}

type StorageConfig struct {
	Driver     string           `yaml:"driver" json:"driver" default:"fs" env:"STORAGE_DRIVER"`
	Filesystem FileSystemConfig `yaml:"fs" json:"fs"`
	S3         S3Config         `yaml:"s3" json:"s3"`
	GCS        GCSConfig        `yaml:"gcs" json:"gcs"`
}

type S3Config struct {
	Hostname        string `yaml:"hostname" json:"hostname" default:"" env:"S3_HOSTNAME"`
	Port            string `yaml:"port" json:"port" default:"" env:"S3_PORT"`
	AccessKeyID     string `yaml:"access_key_id" json:"access_key_id" env:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `yaml:"secret_access_key" json:"secret_access_key" env:"S3_SECRET_ACCESS_KEY"`
	Region          string `yaml:"region" json:"region" env:"S3_REGION"`
	Bucket          string `yaml:"bucket" json:"bucket" env:"S3_BUCKET"`
	EnableSSL       bool   `yaml:"enable_ssl" json:"enable_ssl" default:"true" env:"S3_ENABLE_SSL"`
	UsePathStyle    bool   `yaml:"use_path_style" json:"use_path_style" default:"false" env:"S3_ENABLE_PATH_STYLE"`
}

type FTPConfig struct {
	Address  string
	Username string
	Password string
}

// Storage is the storage interface.
type Storage interface {
	Save(ctx context.Context, content io.Reader, path string) error
	Stat(ctx context.Context, path string) (*Stat, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

// Stat contains metadata about content stored in storage.
type Stat struct {
	ModifiedTime time.Time
	Size         int64
}

// ErrNotExist is a sentinel error returned by the Open and the Stat methods.
var ErrNotExist = errors.New("file does not exist")
