package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	storageApi "github.com/mediaprodcast/commons/api/storage"
	"github.com/mediaprodcast/storage/pkg/storage"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"go.uber.org/zap"
)

const (
	maxUploadSize        = 50 * 1024 * 1024 // 50MB
	defaultUploadWorkers = 1
	defaultDeleteWorkers = 1
)

var (
	ErrPoolBusy         = errors.New("server busy, please try again later")
	ErrInvalidPath      = errors.New("invalid path")
	ErrInvalidExtension = errors.New("invalid file extension")
)

// WorkerPool manages concurrent task execution with limited workers
type WorkerPool struct {
	taskChan  chan func() error
	waitGroup sync.WaitGroup
}

// NewWorkerPool creates a new pool with specified worker count
func NewWorkerPool(maxWorkers int) *WorkerPool {
	pool := &WorkerPool{
		taskChan: make(chan func() error, maxWorkers*2),
	}
	pool.waitGroup.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go pool.worker()
	}
	return pool
}

func (p *WorkerPool) worker() {
	defer p.waitGroup.Done()
	for task := range p.taskChan {
		if err := task(); err != nil {
			// Log errors through your preferred logging mechanism
		}
	}
}

// Submit task to the pool with timeout from context
func (p *WorkerPool) Submit(ctx context.Context, task func() error) error {
	select {
	case p.taskChan <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrPoolBusy
	}
}

// StorageServer handles storage requests with worker pools
type StorageServer struct {
	uploadPool   *WorkerPool
	deletePool   *WorkerPool
	stopChan     chan struct{}
	shutdownOnce sync.Once
}

// NewStorageHandler creates a new server instance
func NewStorageHandler(uploadWorkers, deleteWorkers int) *StorageServer {
	if uploadWorkers <= 0 {
		uploadWorkers = defaultUploadWorkers
	}
	if deleteWorkers <= 0 {
		deleteWorkers = defaultDeleteWorkers
	}

	return &StorageServer{
		uploadPool: NewWorkerPool(uploadWorkers),
		deletePool: NewWorkerPool(deleteWorkers),
		stopChan:   make(chan struct{}),
	}
}

// RequestHandler routes incoming requests
func (s *StorageServer) RequestHandler(w http.ResponseWriter, r *http.Request) {
	logger := zap.L().With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	token, err := getRequestAccessToken(r)
	if err != nil {
		writeError(logger, w, "Unauthorized: invalid credentials", http.StatusUnauthorized, err)
		return
	}

	cfg, err := storageApi.DecodeStorageConfig(token)
	if err != nil {
		writeError(logger, w, "Invalid storage config", http.StatusUnauthorized, err)
		return
	}

	driver, err := storage.NewStorage(cfg)
	if err != nil {
		writeError(logger, w, "Storage init failed", http.StatusInternalServerError, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleServe(r.Context(), driver, w, r, logger)
	case http.MethodPost, http.MethodPut:
		s.handleUpload(r.Context(), driver, w, r, logger)
	case http.MethodDelete:
		s.handleDelete(r.Context(), driver, w, r, logger)
	default:
		writeError(logger, w, "Method not allowed", http.StatusMethodNotAllowed, nil)
	}
}

// Shutdown gracefully stops the server
func (s *StorageServer) Shutdown() {
	s.shutdownOnce.Do(func() {
		close(s.stopChan)
		s.uploadPool.waitGroup.Wait()
		s.deletePool.waitGroup.Wait()
	})
}

func (s *StorageServer) handleUpload(
	ctx context.Context,
	driver defs.Storage,
	w http.ResponseWriter,
	r *http.Request,
	logger *zap.Logger,
) {
	path, err := validateUploadPath(r.URL.Path)
	if err != nil {
		writeError(logger, w, "Invalid upload path", http.StatusBadRequest, err)
		return
	}

	contentType := determineContentType(r.Header.Get("Content-Type"), path)
	logger.Debug("Processing upload", zap.String("path", path), zap.String("content_type", contentType))

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	defer r.Body.Close()

	// uploadTask := func() error {
	// 	return driver.Save(ctx, r.Body, path)
	// }

	// if err := s.uploadPool.Submit(ctx, uploadTask); err != nil {
	// 	status := http.StatusInternalServerError
	// 	if errors.Is(err, defs.ErrNotExist) {
	// 		status = http.StatusBadRequest
	// 	} else if errors.Is(err, ErrPoolBusy) {
	// 		status = http.StatusTooManyRequests
	// 	}
	// 	writeError(logger,  w, "Upload failed", status, err)
	// 	return
	// }

	if err := driver.Save(ctx, r.Body, path); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, defs.ErrNotExist) {
			status = http.StatusBadRequest
		} else if errors.Is(err, ErrPoolBusy) {
			status = http.StatusTooManyRequests
		}
		writeError(logger, w, "Upload failed", status, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *StorageServer) handleDelete(
	ctx context.Context,
	driver defs.Storage,
	w http.ResponseWriter,
	r *http.Request,
	logger *zap.Logger,
) {
	path, err := sanitizePath(r.URL.Path)
	if err != nil {
		writeError(logger, w, "Invalid delete path", http.StatusBadRequest, err)
		return
	}

	deleteTask := func() error {
		return driver.Delete(ctx, path)
	}

	if err := s.deletePool.Submit(ctx, deleteTask); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, defs.ErrNotExist) {
			status = http.StatusNotFound
		} else if errors.Is(err, ErrPoolBusy) {
			status = http.StatusTooManyRequests
		}
		writeError(logger, w, "Delete failed", status, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleServe processes serve requests
func (s *StorageServer) handleServe(
	ctx context.Context,
	driver defs.Storage,
	w http.ResponseWriter, // The response writer
	r *http.Request, // The request
	logger *zap.Logger, // The logger
) {
	// Validate and extract path
	path, err := sanitizePath(r.URL.Path)
	if err != nil {
		writeError(logger, w, "Invalid file path", http.StatusBadRequest, err)
		return
	}

	logger.Debug("Processing serve", zap.String("path", path))

	// Check if the path is a file or directory
	// if it's a file, serve the file
	ext := filepath.Ext(path)
	if ext != "" {
		contentType := mime.TypeByExtension(ext)
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		// Open the file
		reader, err := driver.Open(ctx, path)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if errors.Is(err, defs.ErrNotExist) {
				statusCode = http.StatusNotFound
			}
			writeError(logger, w, "File open failed", statusCode, err)
		}

		// Copy the file to the response
		if _, err := io.Copy(w, reader); err != nil {
			writeError(logger, w, "Failed to write file to response", http.StatusInternalServerError, err)
		}
		return
	}

	// List directory
	// if it's a directory, list the files
	files, err := driver.List(ctx, path)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, defs.ErrNotExist) {
			statusCode = http.StatusNotFound
		}
		writeError(logger, w, "File list failed", statusCode, err)
		return
	}

	// Write file list to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(files); err != nil {
		writeError(logger, w, "Failed to encode response", http.StatusInternalServerError, err)
	}
}

// Validation helpers
func validateUploadPath(rawPath string) (string, error) {
	path, err := sanitizePath(rawPath)
	if err != nil {
		return "", err
	}

	if ext := filepath.Ext(path); ext == "" {
		return "", ErrInvalidExtension
	}

	return path, nil
}

func determineContentType(headerType, path string) string {
	if headerType != "" {
		return headerType
	}
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// writeError handles error reporting consistently
func writeError(logger *zap.Logger, w http.ResponseWriter, message string, statusCode int, err error) {
	// Log the error with details
	logger.Error(message,
		zap.Error(err),
		zap.Int("status_code", statusCode),
	)

	// Prepare client response
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, message)
}

// getRequestAccessToken retrieves the storage configuration from the Authorization header
func getRequestAccessToken(r *http.Request) (string, error) {
	token, err := parseBearerToken(r)
	if err != nil {
		// return "", fmt.Errorf("authorization failed: %w", err)
		// TODO: fallback to user again to support shaka packager since it only sends one header (user-agent)
		return r.UserAgent(), nil
	}

	return token, nil
}

// sanitizePath validates and cleans the request path
func sanitizePath(urlPath string) (string, error) {
	path := strings.TrimPrefix(urlPath, "/")
	if path == "" {
		return "", errors.New("empty path")
	}

	// Prevent path traversal attacks
	if strings.Contains(path, "..") {
		return "", errors.New("invalid path contains traversal characters")
	}

	// Clean the path to remove any redundant elements
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return "", fmt.Errorf("invalid path format: %s", path)
	}

	return cleanPath, nil
}

// parseBearerToken extracts the Bearer token from the Authorization header
func parseBearerToken(r *http.Request) (string, error) {
	token := r.URL.Query().Get("token")
	if token != "" {
		return token, nil
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}
