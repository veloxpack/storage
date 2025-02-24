package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mediaprodcast/storage/pkg/backend"
	"github.com/mediaprodcast/storage/pkg/backend/server"
	"github.com/mediaprodcast/storage/pkg/storage"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	be := backend.NewStorageBackend(
		storage.WithDriver(os.Getenv("STORAGE_DRIVER")),
		storage.WithOutputLocation(os.Getenv("STORAGE_OUTPUT_LOCATION")),
	)

	storageServer, err := be.Server(
		server.WithLogger(logger),
		server.WithHTTPAddr(os.Getenv("STORAGE_ADDR")),
		server.WithDeletePoolSize(5),
		server.WithUploadPoolSize(5),
	)
	if err != nil {
		logger.Fatal("failed to storage backend server", zap.Error(err))
	}

	// Log the actual address
	logger.Info("Server started", zap.String("address", storageServer.Addr))

	// Graceful shutdown setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		if err := storageServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Signal channel for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutting down server gracefully...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := storageServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}
}
