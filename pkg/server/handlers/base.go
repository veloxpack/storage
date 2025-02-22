package handlers

import (
	"context"
	"fmt"
	"net/http"

	storageApi "github.com/mediaprodcast/commons/api/storage"
	"github.com/mediaprodcast/storage/pkg/server/utils"
	"github.com/mediaprodcast/storage/pkg/server/worker"
	"github.com/mediaprodcast/storage/pkg/storage"
	defs "github.com/mediaprodcast/storage/pkg/storage/defs"
)

type StorageHandler struct {
	upload    *UploadHandler
	download  *DownloadHandler
	delete    *DeleteHandler
	streaming *StreamingHandler
}

func NewStorageHandler(uploadPool, deletePool *worker.Pool) *StorageHandler {
	streaming := NewStreamingHandler()
	return &StorageHandler{
		streaming: streaming,
		upload:    NewUploadHandler(uploadPool, utils.MaxUploadSize, streaming),
		download:  NewDownloadHandler(streaming),
		delete:    NewDeleteHandler(deletePool),
	}
}

func (h *StorageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	storageBackend, err := h.getStoragestorageBackend(ctx)
	if err != nil {
		utils.WriteError(w, "Storage init failed", http.StatusInternalServerError, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.download.Handle(ctx, storageBackend, w, r)
	case http.MethodPost, http.MethodPut:
		h.upload.Handle(ctx, storageBackend, w, r)
	case http.MethodDelete:
		h.delete.Handle(ctx, storageBackend, w, r)
	default:
		utils.WriteError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
	}
}

func (h *StorageHandler) getStoragestorageBackend(ctx context.Context) (defs.Storage, error) {
	cfg, err := storageApi.DecodeStorageConfig("token")
	if err != nil {
		return nil, fmt.Errorf("invalid storage config: %w", err)
	}
	return storage.NewStorage(cfg)
}

func (h *StorageHandler) Shutdown() {
	h.streaming.Shutdown()
}
