package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/mediaprodcast/storage/pkg/server/utils"
	"github.com/mediaprodcast/storage/pkg/server/worker"
	"github.com/mediaprodcast/storage/pkg/storage"
	"github.com/mediaprodcast/storage/pkg/storage/provider"
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
	storageBackend, err := h.getStorageBackend(ctx)
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

func (h *StorageHandler) getStorageBackend(ctx context.Context) (provider.Storage, error) {
	return storage.NewStorage(
		storage.WithDriver(os.Getenv("STORAGE_DRIVER")),
		storage.WithOutputLocation(os.Getenv("STORAGE_OUTPUT_LOCATION")),
	)
}

func (h *StorageHandler) Shutdown() {
	h.streaming.Shutdown()
}
