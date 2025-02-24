package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/veloxpack/storage/pkg/backend/server/middleware"
	"github.com/veloxpack/storage/pkg/backend/server/utils"
	"github.com/veloxpack/storage/pkg/storage/provider"
	"go.uber.org/zap"
)

type DownloadHandler struct {
	logger    *zap.Logger
	streaming *StreamingHandler
}

func NewDownloadHandler(streaming *StreamingHandler) *DownloadHandler {
	return &DownloadHandler{
		logger:    zap.L().Named("download"),
		streaming: streaming,
	}
}

func (h *DownloadHandler) Handle(ctx context.Context, storageBackend provider.Storage, w http.ResponseWriter, r *http.Request) {
	path := middleware.GetValidatedPath(ctx)

	if au, exists := h.streaming.GetActiveUpload(path); exists {
		h.serveActiveUpload(w, r, au)
		return
	}

	if filepath.Ext(path) != "" {
		h.serveFile(ctx, storageBackend, w, path)
		return
	}

	h.listFiles(ctx, storageBackend, w, path)
}

func (h *DownloadHandler) serveFile(ctx context.Context, storageBackend provider.Storage, w http.ResponseWriter, path string) {
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	reader, err := storageBackend.Open(ctx, path)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, provider.ErrNotExist) {
			status = http.StatusNotFound
		}
		utils.WriteError(w, "File open failed", status, err)
		return
	}
	defer reader.Close()

	if _, err := io.Copy(w, reader); err != nil {
		h.logger.Error("Failed to stream file", zap.Error(err))
	}
}

func (h *DownloadHandler) listFiles(ctx context.Context, storageBackend provider.Storage, w http.ResponseWriter, path string) {
	files, err := storageBackend.List(ctx, path)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, provider.ErrNotExist) {
			status = http.StatusNotFound
		}
		utils.WriteError(w, "List files failed", status, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		zap.L().Error("Failed to encode response", zap.Error(err))
	}
}

func (h *DownloadHandler) serveActiveUpload(w http.ResponseWriter, r *http.Request, au *ActiveUpload) {
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", utils.DetermineContentType(au.header.Get("Content-Type"), r.URL.Path))
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.WriteError(w, "Streaming unsupported", http.StatusInternalServerError, nil)
		return
	}

	var offset int
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			au.mu.RLock()
			if offset >= len(au.buffer) && au.eof {
				au.mu.RUnlock()
				return
			}

			if offset < len(au.buffer) {
				n, err := w.Write(au.buffer[offset:])
				if err != nil {
					au.mu.RUnlock()
					return
				}
				offset += n
				flusher.Flush()
			}
			au.mu.RUnlock()

		case <-r.Context().Done():
			return
		}
	}
}
