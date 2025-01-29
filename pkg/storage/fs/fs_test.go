package fs

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"github.com/stretchr/testify/assert"
)

func removeDir(path string) {
	_ = os.RemoveAll(path)
}

func TestFS(t *testing.T) {
	cfg := Config{
		Root: "./tmp",
	}

	t.Run("should return error file does not exist", func(t *testing.T) {
		s := NewStorage(cfg)
		defer removeDir(cfg.Root)

		ctx := context.Background()

		_, err := s.Stat(ctx, "doesnotexist")
		assert.EqualError(t, err, defs.ErrNotExist.Error())
	})

	t.Run("should create file", func(t *testing.T) {
		s := NewStorage(cfg)
		defer removeDir(cfg.Root)

		ctx := context.Background()

		err := s.Save(ctx, bytes.NewBufferString("hello"), "world")
		assert.NoError(t, err)
	})

	t.Run("should get metadata of file", func(t *testing.T) {
		s := NewStorage(cfg)
		defer removeDir(cfg.Root)

		ctx := context.Background()

		before := time.Now().Add(-1 * time.Second)

		err := s.Save(ctx, bytes.NewBufferString("hello"), "world")
		assert.NoError(t, err)

		now := time.Now().Add(2 * time.Second)

		stat, err := s.Stat(ctx, "world")
		assert.NoError(t, err)

		assert.Equal(t, int64(5), stat.Size)
		assert.False(t, stat.ModifiedTime.Before(before))
		assert.False(t, stat.ModifiedTime.After(now))
	})

	t.Run("should create then delete file", func(t *testing.T) {
		s := NewStorage(cfg)
		defer removeDir(cfg.Root)

		ctx := context.Background()

		err := s.Save(ctx, bytes.NewBufferString("hello"), "world")
		assert.NoError(t, err)

		err = s.Delete(ctx, "world")
		assert.NoError(t, err)

		_, err = s.Stat(ctx, "world")
		assert.EqualError(t, err, defs.ErrNotExist.Error())
	})

	t.Run("should create then open file", func(t *testing.T) {
		s := NewStorage(cfg)
		defer removeDir(cfg.Root)

		ctx := context.Background()

		err := s.Save(ctx, bytes.NewBufferString("hello"), "world")
		assert.NoError(t, err)

		f, err := s.Open(ctx, "world")
		assert.NoError(t, err)
		defer f.Close()

		b, err := io.ReadAll(f) // Updated to use io.ReadAll
		assert.NoError(t, err)
		assert.Equal(t, "hello", string(b))
	})

	t.Run("should try to open a non existing file", func(t *testing.T) {
		s := NewStorage(cfg)
		defer removeDir(cfg.Root)

		ctx := context.Background()

		_, err := s.Open(ctx, "world")
		assert.EqualError(t, err, defs.ErrNotExist.Error())
	})
}
