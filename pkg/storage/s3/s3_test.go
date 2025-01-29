package s3_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/mediaprodcast/storage/pkg/storage"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/assert"
)

func TestS3(t *testing.T) {
	// Setting up the fake minio server
	s3accessKey := "access_key"
	s3secretKey := "secret_key"
	s3bucket := "mpc-aws-bucket"

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	dockerHostPort := "9000"
	options := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9000/tcp": {
				{
					HostPort: dockerHostPort,
				},
			},
		},
		Env: []string{"MINIO_ACCESS_KEY=" + s3accessKey, "MINIO_SECRET_KEY=" + s3secretKey},
	}

	resource, err := pool.RunWithOptions(options, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	endpoint := fmt.Sprintf("localhost:%s", resource.GetPort(dockerHostPort+"/tcp"))

	if err := pool.Retry(func() error {
		url := fmt.Sprintf("http://%s/minio/health/live", endpoint)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New("status code not OK")
		}
		return nil
	}); err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	defer func(t *testing.T) {
		if err = pool.Purge(resource); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	}(t)

	cfg := &defs.StorageConfig{
		S3: defs.S3Config{
			Hostname:        endpoint,
			AccessKeyID:     s3accessKey,
			SecretAccessKey: s3secretKey,
			Bucket:          s3bucket,
			EnableSSL:       false,
		},
	}

	t.Run("should return error file does not exist", func(t *testing.T) {
		s, err := storage.NewStorage(defs.AmazonS3, cfg)
		assert.NoError(t, err)

		ctx := context.Background()

		_, err = s.Stat(ctx, "doesnotexist")
		assert.EqualError(t, err, "The specified key does not exist.")
	})

	t.Run("should create file", func(t *testing.T) {
		s, err := storage.NewStorage(defs.AmazonS3, cfg)
		assert.NoError(t, err)

		ctx := context.Background()

		err = s.Save(ctx, bytes.NewBufferString("hello"), "world")
		assert.NoError(t, err)
	})

	t.Run("should get metadata of file", func(t *testing.T) {
		s, err := storage.NewStorage(defs.AmazonS3, cfg)
		assert.NoError(t, err)

		ctx := context.Background()

		before := time.Now().Add(-1 * time.Second)

		path := "hello/world/master.m3u8"

		err = s.Save(ctx, bytes.NewBufferString("hello"), path)
		assert.NoError(t, err)

		now := time.Now().Add(2 * time.Second)

		stat, err := s.Stat(ctx, path)
		assert.NoError(t, err)

		assert.Equal(t, int64(5), stat.Size)
		assert.Equal(t, false, stat.ModifiedTime.Before(before))
		assert.Equal(t, false, stat.ModifiedTime.After(now))
	})

	t.Run("should create then delete file", func(t *testing.T) {
		s, err := storage.NewStorage(defs.AmazonS3, cfg)
		assert.NoError(t, err)

		ctx := context.Background()

		err = s.Save(ctx, bytes.NewBufferString("hello"), "world")
		assert.NoError(t, err)

		err = s.Delete(ctx, "world")
		assert.NoError(t, err)

		_, err = s.Stat(ctx, "world")
		assert.EqualError(t, err, "The specified key does not exist.")
	})

	t.Run("should create then open file", func(t *testing.T) {
		s, err := storage.NewStorage(defs.AmazonS3, cfg)
		assert.NoError(t, err)

		ctx := context.Background()

		err = s.Save(ctx, bytes.NewBufferString("hello"), "world")
		assert.NoError(t, err)

		f, err := s.Open(ctx, "world")
		assert.NoError(t, err)
		defer func() { _ = f.Close() }()

		b, err := io.ReadAll(f)
		assert.NoError(t, err)
		assert.Equal(t, "hello", string(b))
	})

	t.Run("should delete the file", func(t *testing.T) {
		s, err := storage.NewStorage(defs.AmazonS3, cfg)
		assert.NoError(t, err)

		ctx := context.Background()

		err = s.Delete(ctx, "world")
		assert.NoError(t, err)
	})
}
