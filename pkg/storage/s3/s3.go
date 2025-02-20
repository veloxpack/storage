package s3

import (
	"context"
	"io"
	"mime"
	"path/filepath"

	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Storage is a S3 storage.
type Storage struct {
	bucket string
	client *minio.Client
}

func NewStorage(cfg *pb.S3Config) (*Storage, error) {
	ctx := context.Background()

	s, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyId, cfg.SecretAccessKey, ""),
		Secure: cfg.EnableSsl,
	})
	if err != nil {
		return nil, err
	}

	found, err := s.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}

	if !found {
		err := s.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region})
		if err != nil {
			return nil, err
		}
	}

	return &Storage{
		client: s,
		bucket: cfg.Bucket,
	}, err
}

// Save saves content to path.
func (s *Storage) Save(ctx context.Context, content io.Reader, path string) error {
	_, err := s.client.PutObject(ctx, s.bucket, path, content, -1, minio.PutObjectOptions{
		ContentType: mime.TypeByExtension(filepath.Ext(path)),
	})

	return err
}

// Stat returns path metadata.
func (s *Storage) Stat(ctx context.Context, path string) (*pb.Stat, error) {
	stat, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	objectStat := &pb.Stat{
		ModifiedTime: timestamppb.New(stat.LastModified),
		Size:         stat.Size,
	}

	return objectStat, nil
}

// Open opens path for reading.
func (s *Storage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	return object, err
}

// Delete deletes path.
func (s *Storage) Delete(ctx context.Context, path string) error {
	err := s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
	return err
}

// List lists path contents.
func (s *Storage) List(ctx context.Context, path string) ([]*pb.Stat, error) {
	var objects []*pb.Stat

	doneCh := make(chan struct{})
	defer close(doneCh)

	for object := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}) {
		if object.Err != nil {
			return nil, object.Err
		}

		objects = append(objects, &pb.Stat{
			ModifiedTime: timestamppb.New(object.LastModified),
			Size:         object.Size,
			Name:         object.Key,
			ContentType:  mime.TypeByExtension(filepath.Ext(object.Key)),
		})
	}

	return objects, nil
}
