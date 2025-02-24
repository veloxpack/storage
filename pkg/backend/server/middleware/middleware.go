package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/mediaprodcast/storage/pkg/backend/server/utils"
	"go.uber.org/zap"
)

// Middleware chaining helper
func ChainMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- { // Apply middlewares in reverse order
		h = middlewares[i](h)
	}
	return h
}

type pathContextKey string

const ValidatedPathContextKey pathContextKey = "validatedPath"

func GetValidatedPath(ctx context.Context) string {
	return ctx.Value(ValidatedPathContextKey).(string)
}

func PathValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, err := utils.SanitizePath(r.URL.Path)
		if err != nil {
			utils.WriteError(w, "Invalid path", http.StatusBadRequest, err)
			return
		}

		ctx := context.WithValue(r.Context(), ValidatedPathContextKey, path)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		// Log request details
		zap.L().Info("HTTP Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.String()),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Duration("duration", time.Since(start)),
			zap.String("content-type", utils.DetermineContentType(r.Header.Get("Content-Type"), r.URL.String())),
			zap.Any("transfer-encoding", r.TransferEncoding),
		)
	})
}
