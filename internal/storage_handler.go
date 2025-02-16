package internal

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	storageApi "github.com/mediaprodcast/commons/api/storage"
	"github.com/mediaprodcast/storage/pkg/storage"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	maxUploadSize = 5 * 1024 * 1024 * 1024 // 5GB
)

// StorageHandler handles incoming storage requests.
func StorageHandler(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("http")
	ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.RequestURI))
	defer span.End()

	logger := zap.L().With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Validate HTTP method
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		// Continue processing
	default:
		writeError(logger, span, w, "Method Not Allowed", http.StatusMethodNotAllowed, nil)
		return
	}

	// Get storage configuration
	token, err := getRequestAccessToken(r)
	if err != nil {
		writeError(logger, span, w, "Unauthorized: invalid or missing credentials", http.StatusUnauthorized, err)
		return
	}

	// Decode storage configuration
	cfg, err := storageApi.DecodeStorageConfig(token)
	if err != nil {
		writeError(logger, span, w, "Invalid storage configuration", http.StatusUnauthorized, err)
		return
	}

	// Initialize storage driver
	driver, err := storage.NewStorage(cfg)
	if err != nil {
		writeError(logger, span, w, "Failed to initialize storage", http.StatusInternalServerError, err)
		return
	}

	// Route requests
	switch r.Method {
	case http.MethodPost, http.MethodPut:
		handleUpload(ctx, driver, w, r, logger)
	case http.MethodDelete:
		handleDelete(ctx, driver, w, r, logger)
	}
}

// writeError handles error reporting consistently
func writeError(logger *zap.Logger, span trace.Span, w http.ResponseWriter, message string, statusCode int, err error) {
	// Set OpenTelemetry status
	span.SetStatus(otelCodes.Error, message)
	if err != nil {
		span.RecordError(err)
	}

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

// handleUpload processes file upload requests
func handleUpload(
	ctx context.Context,
	driver defs.Storage,
	w http.ResponseWriter,
	r *http.Request,
	logger *zap.Logger,
) {
	span := trace.SpanFromContext(ctx)

	// Validate and extract path
	path, err := sanitizePath(r.URL.Path)
	if err != nil {
		writeError(logger, span, w, "Invalid file path", http.StatusBadRequest, err)
		return
	}

	// Determine content type
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(path))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	logger.Info("Processing upload",
		zap.String("path", path),
		zap.String("content_type", contentType),
	)

	// Limit upload size and ensure body closure
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	defer r.Body.Close()

	// Save to storage
	if err := driver.Save(ctx, r.Body, path); err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, defs.ErrNotExist) {
			statusCode = http.StatusBadRequest
		}

		writeError(logger, span, w, "File upload failed", statusCode, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleDelete processes file deletion requests
func handleDelete(
	ctx context.Context,
	driver defs.Storage,
	w http.ResponseWriter,
	r *http.Request,
	logger *zap.Logger,
) {
	span := trace.SpanFromContext(ctx)

	// Validate and extract path
	path, err := sanitizePath(r.URL.Path)
	if err != nil {
		writeError(logger, span, w, "Invalid file path", http.StatusBadRequest, err)
		return
	}

	logger.Info("Processing delete", zap.String("path", path))

	if err := driver.Delete(ctx, path); err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, defs.ErrNotExist) {
			statusCode = http.StatusNotFound
		}
		writeError(logger, span, w, "File deletion failed", statusCode, err)
		return
	}

	w.WriteHeader(http.StatusOK)
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
