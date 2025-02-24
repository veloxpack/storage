package handlers

import (
	"net/http"

	"github.com/mediaprodcast/storage/pkg/backend/server/utils"
	"github.com/mediaprodcast/storage/pkg/backend/server/worker"
	"github.com/mediaprodcast/storage/pkg/storage/provider"
)

type StorageHandler struct {
	upload    *UploadHandler
	download  *DownloadHandler
	delete    *DeleteHandler
	streaming *StreamingHandler
	storage   provider.Storage
}

func NewStorageHandler(storage provider.Storage, uploadPool, deletePool *worker.Pool) *StorageHandler {
	streaming := NewStreamingHandler()

	return &StorageHandler{
		storage:   storage,
		streaming: streaming,
		upload:    NewUploadHandler(uploadPool, utils.MaxUploadSize, streaming),
		download:  NewDownloadHandler(streaming),
		delete:    NewDeleteHandler(deletePool),
	}
}

func (h *StorageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.download.Handle(ctx, h.storage, w, r)
	case http.MethodPost, http.MethodPut:
		h.upload.Handle(ctx, h.storage, w, r)
	case http.MethodDelete:
		h.delete.Handle(ctx, h.storage, w, r)
	default:
		utils.WriteError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
	}
}

func (h *StorageHandler) Shutdown() {
	h.streaming.Shutdown()
}
