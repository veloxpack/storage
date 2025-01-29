package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mediaprodcast/commons/discovery"
	"github.com/mediaprodcast/commons/env"
	"github.com/mediaprodcast/commons/tracer"
	"github.com/mediaprodcast/storage/internal"
	"go.uber.org/zap"
)

var httpAddr = env.GetString("HTTP_ADDR", ":9500")

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	// Set up global tracing
	if err := tracer.SetGlobalTracer(context.TODO(), discovery.StorageSvsName); err != nil {
		logger.Fatal("Could not set global tracer", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Register service in Consul
	instanceID, registry, err := discovery.Register(ctx, discovery.StorageSvsName, httpAddr)
	if err != nil {
		logger.Error("Failed to register service", zap.Error(err))
	}
	defer registry.Deregister(ctx, instanceID, discovery.StorageSvsName)

	mux := http.NewServeMux()

	// Storage proxy
	mux.HandleFunc("/", internal.StorageHandler)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	// Run server in a goroutine
	go func() {
		logger.Info("Starting storage service", zap.String("port", httpAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Failed to shut down server gracefully", zap.Error(err))
	} else {
		logger.Info("Server shut down successfully")
	}
}
