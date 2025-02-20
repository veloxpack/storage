package gcs

import (
	"context"
	"io"
	"mime"
	"path/filepath"

	gstorage "cloud.google.com/go/storage"
	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Storage is a gcs storage.
type Storage struct {
	bucket *gstorage.BucketHandle
}

// NewStorage returns a new Storage.
func NewStorage(ctx context.Context, credentialsFile, bucket string) (*Storage, error) {
	client, err := gstorage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}

	return &Storage{bucket: client.Bucket(bucket)}, nil
}

// Save saves content to path.
func (g *Storage) Save(ctx context.Context, content io.Reader, path string) (rerr error) {
	w := g.bucket.Object(path).NewWriter(ctx)
	w.ContentType = mime.TypeByExtension(filepath.Ext(path))

	defer func() {
		if err := w.Close(); err != nil {
			rerr = err
		}
	}()

	if _, err := io.Copy(w, content); err != nil {
		return err
	}

	return rerr
}

// Stat returns path metadata.
func (g *Storage) Stat(ctx context.Context, path string) (*pb.Stat, error) {
	attrs, err := g.bucket.Object(path).Attrs(ctx)
	if err == gstorage.ErrObjectNotExist {
		return nil, defs.ErrNotExist
	} else if err != nil {
		return nil, err
	}

	return &pb.Stat{
		ModifiedTime: timestamppb.New(attrs.Updated),
		Size:         attrs.Size,
	}, nil
}

// Open opens path for reading.
func (g *Storage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	r, err := g.bucket.Object(path).NewReader(ctx)
	if err == gstorage.ErrObjectNotExist {
		return nil, defs.ErrNotExist
	}
	return r, err
}

// Delete deletes path.
func (g *Storage) Delete(ctx context.Context, path string) error {
	return g.bucket.Object(path).Delete(ctx)
}

// List lists path contents.
func (g *Storage) List(ctx context.Context, path string) ([]*pb.Stat, error) {
	it := g.bucket.Objects(ctx, &gstorage.Query{Prefix: path})
	var stats []*pb.Stat
	for {
		attrs, err := it.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		stats = append(stats, &pb.Stat{
			ModifiedTime: timestamppb.New(attrs.Updated),
			Size:         attrs.Size,
			Name:         attrs.Name,
			ContentType:  attrs.ContentType,
		})
	}

	return stats, nil
}
