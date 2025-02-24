package handlers

import (
	"context"
	"net/http"

	"github.com/veloxpack/storage/pkg/backend/server/middleware"
	"github.com/veloxpack/storage/pkg/backend/server/utils"
	"github.com/veloxpack/storage/pkg/backend/server/worker"
	"github.com/veloxpack/storage/pkg/storage/provider"
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

func (h *DeleteHandler) Handle(ctx context.Context, storageBackend provider.Storage, w http.ResponseWriter, r *http.Request) {
	path := middleware.GetValidatedPath(ctx)

	task := func() {
		if err := storageBackend.Delete(context.Background(), path); err != nil {
			h.logger.Error("Delete failed", zap.String("path", path), zap.Error(err))
		}
	}

	if err := h.pool.Submit(task); err != nil {
		utils.WriteError(w, "Delete failed to submit", http.StatusTooManyRequests, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
