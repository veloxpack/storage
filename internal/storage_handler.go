package internal

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/mediaprodcast/storage/pkg/storage"
	"github.com/mediaprodcast/storage/pkg/storage/defs"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// StorageHandler handles incoming storage requests, determining the storage configuration
// from either the User-Agent header or the X-Storage-Config header.
func StorageHandler(w http.ResponseWriter, r *http.Request) {
	// Create a tracer span
	tr := otel.Tracer("http")
	ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.RequestURI))
	defer span.End()

	cfgStr := getStorageConfig(r)
	if cfgStr == "" {
		http.Error(w, "missing storage configuration", http.StatusUnprocessableEntity)
		return
	}

	cfg, err := storage.Decode(cfgStr)
	if err != nil {
		http.Error(w, "invalid storage configuration", http.StatusUnprocessableEntity)
		return
	}

	driver, err := storage.NewStorage(cfg)
	if err != nil {
		http.Error(w, "failed to initialize storage", http.StatusUnprocessableEntity)
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

// getStorageConfig retrieves the storage configuration from the request headers.
func getStorageConfig(r *http.Request) string {
	cfgStr := r.Header.Get("X-Storage-Config")
	if cfgStr == "" {
		cfgStr = r.UserAgent()
	}
	return cfgStr
}

// handleUpload processes file upload requests.
// It extracts the file path, determines its MIME type, and saves the content.
func handleUpload(ctx context.Context, driver defs.Storage, w http.ResponseWriter, r *http.Request) {
	path := extractPath(r.URL.Path)
	contentType := mime.TypeByExtension(filepath.Ext(path))

	zap.L().Info("Uploading file", zap.String("path", path), zap.String("contentType", contentType))

	if err := driver.Save(ctx, r.Body, path); err != nil {
		http.Error(w, "Failed to upload file", http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleDelete processes file deletion requests.
// It extracts the file path from the request URL and deletes the corresponding file.
func handleDelete(ctx context.Context, driver defs.Storage, w http.ResponseWriter, r *http.Request) {
	path := extractPath(r.URL.Path)
	zap.L().Info("Deleting file", zap.String("path", path))

	if err := driver.Delete(ctx, path); err != nil {
		http.Error(w, "Failed to delete file", http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// extractPath extracts the file path from the request URL.
func extractPath(urlPath string) string {
	return strings.TrimPrefix(urlPath, "/")
}
