package internal

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/mediaprodcast/storage/pkg/credentials"
	"github.com/mediaprodcast/storage/pkg/storage"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// StorageHandler handles incoming storage requests.
func StorageHandler(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("http")
	ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.RequestURI))
	defer span.End()

	cfgStr := getStorageConfig(r)
	if cfgStr == "" {
		writeError(span, w, "missing storage configuration", http.StatusUnprocessableEntity)
		return
	}

	cfg, err := credentials.Decode(cfgStr)
	if err != nil {
		writeError(span, w, "invalid storage configuration", http.StatusUnprocessableEntity)
		return
	}

	driver, err := storage.NewStorage(cfg)
	if err != nil {
		writeError(span, w, "failed to initialize storage", http.StatusUnprocessableEntity)
		return
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		handleUpload(ctx, driver, w, r)
	case http.MethodDelete:
		handleDelete(ctx, driver, w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// writeError logs the error, sets the span status, and writes the HTTP response.
func writeError(span trace.Span, w http.ResponseWriter, message string, statusCode int) {
	span.SetStatus(otelCodes.Error, message)
	zap.L().Error(message)
	http.Error(w, message, statusCode)
}

// getStorageConfig retrieves the storage configuration from the request headers.
func getStorageConfig(r *http.Request) string {
	cfgStr := r.Header.Get("X-Storage-Config")
	if cfgStr == "" {
		cfgStr = r.UserAgent()
	}
	return cfgStr
}

// handleUpload processes file upload requests.
func handleUpload(ctx context.Context, driver defs.Storage, w http.ResponseWriter, r *http.Request) {
	path := extractPath(r.URL.Path)
	contentType := mime.TypeByExtension(filepath.Ext(path))

	zap.L().Info("Uploading file", zap.String("path", path), zap.String("contentType", contentType))

	if err := driver.Save(ctx, r.Body, path); err != nil {
		writeError(trace.SpanFromContext(ctx), w, "Failed to upload file", http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleDelete processes file deletion requests.
func handleDelete(ctx context.Context, driver defs.Storage, w http.ResponseWriter, r *http.Request) {
	path := extractPath(r.URL.Path)
	zap.L().Info("Deleting file", zap.String("path", path))

	if err := driver.Delete(ctx, path); err != nil {
		writeError(trace.SpanFromContext(ctx), w, "Failed to delete file", http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// extractPath extracts the file path from the request URL.
func extractPath(urlPath string) string {
	return strings.TrimPrefix(urlPath, "/")
}
