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
	DataPath string `yaml:"dataPath" json:"dataPath" default:"/data"`
}

type GCSConfig struct {
	CredentialsFile string `yaml:"credentialsFile" json:"credentialsFile" default:""`
	Bucket          string `yaml:"bucket" json:"bucket" default:""`
}

type StorageConfig struct {
	Driver     StorageDriver    `yaml:"driver" json:"driver" default:"fs"`
	Filesystem FileSystemConfig `yaml:"fs" json:"fs"`
	S3         S3Config         `yaml:"s3" json:"s3"`
	GCS        GCSConfig        `yaml:"gcs" json:"gcs"`
}

type S3Config struct {
	Endpoint        string `yaml:"endpoint" json:"endpoint" default:""`
	AccessKeyID     string `yaml:"accessKeyId" json:"accessKeyId"`
	SecretAccessKey string `yaml:"secretAccessKey" json:"secretAccessKey"`
	Region          string `yaml:"region" json:"region"`
	Bucket          string `yaml:"bucket" json:"bucket"`
	EnableSSL       bool   `yaml:"enableSSL" json:"enableSSL" default:"true"`
	UsePathStyle    bool   `yaml:"usePathStyle" json:"usePathStyle" default:"false"`
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
