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
	"github.com/rs/cors"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	mux := http.NewServeMux()

	// Storage proxy
	mux.HandleFunc("/", internal.StorageHandler)

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	// Wrap the handler with CORS
	handlerWithCORS := c.Handler(mux)

	// HTTP2 Setup using h2c
	server := &http.Server{
		Addr:    httpAddr,
		Handler: h2c.NewHandler(handlerWithCORS, &http2.Server{}),
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
