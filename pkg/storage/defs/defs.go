package defs

import (
	"context"
	"errors"
	"io"

	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
)

// Storage is the storage interface.
type Storage interface {
	Save(ctx context.Context, content io.Reader, path string) error
	Stat(ctx context.Context, path string) (*pb.Stat, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

// ErrNotExist is a sentinel error returned by the Open and the Stat methods.
var ErrNotExist = errors.New("file does not exist")
