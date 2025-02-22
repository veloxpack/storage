package handlers

import (
	"context"
	"net/http"

	"github.com/mediaprodcast/storage/pkg/server/middleware"
	"github.com/mediaprodcast/storage/pkg/server/utils"
	"github.com/mediaprodcast/storage/pkg/server/worker"
	defs "github.com/mediaprodcast/storage/pkg/storage/defs"
	"go.uber.org/zap"
)

type DeleteHandler struct {
	pool   *worker.Pool
	logger *zap.Logger
}

func NewDeleteHandler(pool *worker.Pool) *DeleteHandler {
	return &DeleteHandler{
		pool:   pool,
		logger: zap.L().Named("delete"),
	}
}

func (h *DeleteHandler) Handle(ctx context.Context, storageBackend defs.Storage, w http.ResponseWriter, r *http.Request) {
	path := middleware.GetValidatedPath(ctx)

	task := func() {
		if err := storageBackend.Delete(ctx, path); err != nil {
			h.logger.Error("Delete failed", zap.String("path", path), zap.Error(err))
		}
	}

	if err := h.pool.Submit(task); err != nil {
		utils.WriteError(w, "Delete failed to submit", http.StatusTooManyRequests, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
