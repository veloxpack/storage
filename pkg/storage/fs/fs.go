package fs

import (
	"context"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/mediaprodcast/storage/pkg/storage/provider"
)

// Storage is a filesystem storage.
type Storage struct {
	root string
}

// Config is the configuration for Storage.
type Config struct {
	Root string
}

// NewStorage returns a new filesystem storage.
func NewStorage(cfg Config) *Storage {
	return &Storage{root: cfg.Root}
}

func (fs *Storage) abs(path string) string {
	return filepath.Join(fs.root, path)
}

// Save saves content to path.
func (fs *Storage) Save(ctx context.Context, content io.Reader, path string) error {
	abs := fs.abs(path)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return err
	}

	w, err := os.Create(abs)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := io.Copy(w, content); err != nil {
		return err
	}
	return nil
}

// Stat returns path metadata.
func (fs *Storage) Stat(ctx context.Context, path string) (*provider.Stat, error) {
	fi, err := os.Stat(fs.abs(path))
	if os.IsNotExist(err) {
		return nil, provider.ErrNotExist
	} else if err != nil {
		return nil, err
	}

	return &provider.Stat{
		ModifiedTime: fi.ModTime(),
		Size:         fi.Size(),
		ContentType:  mime.TypeByExtension(filepath.Ext(fi.Name())),
		Name:         fi.Name(),
	}, nil
}

// Open opens path for reading.
func (fs *Storage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	f, err := os.Open(fs.abs(path))
	if os.IsNotExist(err) {
		return nil, provider.ErrNotExist
	}
	return f, err
}

// Delete deletes path.
func (fs *Storage) Delete(ctx context.Context, path string) error {
	return os.Remove(fs.abs(path))
}

// List lists path contents.
func (fs *Storage) List(ctx context.Context, path string) ([]*provider.Stat, error) {
	abs := fs.abs(path)
	f, err := os.Open(abs)
	if os.IsNotExist(err) {
		return nil, provider.ErrNotExist
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	stats := make([]*provider.Stat, 0, len(fis))
	for _, fi := range fis {
		stats = append(stats, &provider.Stat{
			Name:         fi.Name(),
			Size:         fi.Size(),
			ModifiedTime: fi.ModTime(),
			ContentType:  mime.TypeByExtension(filepath.Ext(fi.Name())),
		})
	}
	return stats, nil
}
