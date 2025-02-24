package rclone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/veloxpack/storage/pkg/storage/provider"
	// _ "github.com/rclone/rclone/backend/all" // import all backends
	_ "github.com/rclone/rclone/backend/s3"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rclone/rclone/fs/walk"
)

type Storage struct {
	remote string
}

func NewStorage(driver, outputLocation string) *Storage {
	return &Storage{
		remote: fmt.Sprintf(":%s:%s", driver, outputLocation),
	}
}

func (r *Storage) Save(ctx context.Context, content io.Reader, path string) error {
	dstFs, err := r.newFs(ctx)
	if err != nil {
		return err
	}

	_, err = operations.Rcat(ctx, dstFs, path, io.NopCloser(content), time.Now(), nil)
	if err != nil {
		return fmt.Errorf("rcat failed: %w", err)
	}

	return nil
}

func (r *Storage) Stat(ctx context.Context, path string) (*provider.Stat, error) {
	dstFs, err := r.newFs(ctx)
	if err != nil {
		return nil, err
	}

	obj, err := dstFs.NewObject(ctx, path)
	if err != nil {
		if errors.Is(err, fs.ErrorObjectNotFound) {
			return nil, provider.ErrNotExist
		}
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	return &provider.Stat{
		ModifiedTime: obj.ModTime(ctx),
		Size:         obj.Size(),
		Name:         obj.Remote(),
		Path:         path,
	}, nil
}

func (r *Storage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	dstFs, err := r.newFs(ctx)
	if err != nil {
		return nil, err
	}

	obj, err := dstFs.NewObject(ctx, path)
	if err != nil {
		if errors.Is(err, fs.ErrorObjectNotFound) {
			return nil, provider.ErrNotExist
		}
		return nil, fmt.Errorf("open failed: %w", err)
	}

	return obj.Open(ctx)
}

func (r *Storage) Delete(ctx context.Context, path string) error {
	dstFs, err := r.newFs(ctx)
	if err != nil {
		return err
	}

	obj, err := dstFs.NewObject(ctx, path)
	if err != nil {
		if errors.Is(err, fs.ErrorObjectNotFound) {
			return provider.ErrNotExist
		}
		return fmt.Errorf("delete failed: %w", err)
	}

	return operations.DeleteFile(ctx, obj)
}

// List lists path contents.
func (r *Storage) List(ctx context.Context, path string) ([]*provider.Stat, error) {
	dstFs, err := r.newFs(ctx)
	if err != nil {
		return nil, err
	}

	var entries fs.DirEntries
	err = walk.ListR(ctx, dstFs, path, true, -1, walk.ListAll, func(e fs.DirEntries) error {
		entries = e
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	var stats []*provider.Stat
	for _, entry := range entries {
		if obj, ok := entry.(fs.Object); ok {
			stats = append(stats, &provider.Stat{
				ModifiedTime: obj.ModTime(ctx),
				Size:         obj.Size(),
				Name:         obj.Remote(),
				Path:         obj.Remote(),
			})
		}
	}

	return stats, nil
}

func (r *Storage) newFs(ctx context.Context) (fs.Fs, error) {
	dstFs, err := fs.NewFs(ctx, r.remote)
	if err != nil {
		return nil, fmt.Errorf("failed to create fs: %w", err)
	}

	return dstFs, nil
}
