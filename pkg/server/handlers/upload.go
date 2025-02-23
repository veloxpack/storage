package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/mediaprodcast/storage/pkg/server/middleware"
	"github.com/mediaprodcast/storage/pkg/server/utils"
	"github.com/mediaprodcast/storage/pkg/server/worker"
	"github.com/mediaprodcast/storage/pkg/storage/provider"
	"go.uber.org/zap"
)

type UploadHandler struct {
	pool      *worker.Pool
	maxSize   int64
	logger    *zap.Logger
	streaming *StreamingHandler
}

func NewUploadHandler(
	pool *worker.Pool,
	maxSize int64,
	streaming *StreamingHandler,
) *UploadHandler {
	return &UploadHandler{
		pool:      pool,
		maxSize:   maxSize,
		streaming: streaming,
		logger:    zap.L().Named("upload"),
	}
}

func (h *UploadHandler) Handle(ctx context.Context, storageBackend provider.Storage, w http.ResponseWriter, r *http.Request) {
	path := middleware.GetValidatedPath(ctx)

	if h.isChunked(r) {
		h.streaming.HandleChunkedUpload(ctx, storageBackend, w, r, path)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, h.maxSize))
	if err != nil {
		utils.WriteError(w, "Failed to read body", http.StatusBadRequest, err)
		return
	}

	task := func() {
		if err := storageBackend.Save(context.Background(), bytes.NewReader(body), path); err != nil {
			h.logger.Error("Upload failed", zap.Error(err))
		}
	}

	if err := h.pool.Submit(task); err != nil {
		utils.WriteError(w, "Server busy", http.StatusTooManyRequests, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *UploadHandler) isChunked(r *http.Request) bool {
	return len(r.TransferEncoding) > 0 && r.TransferEncoding[0] == "chunked"
}
