package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mediaprodcast/storage/pkg/server/handlers"
	"github.com/mediaprodcast/storage/pkg/server/middleware"
	"github.com/mediaprodcast/storage/pkg/server/worker"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// ServerOption defines a functional option for configuring the server.
type ServerOption func(*ServerConfig)

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	HTTPAddr        string
	Logger          *zap.Logger
	UploadPoolSize  int
	DeletePoolSize  int
	ShutdownTimeout time.Duration
}

// WithHTTPAddr sets the HTTP address for the server.
func WithHTTPAddr(addr string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.HTTPAddr = addr
	}
}

// WithLogger sets the logger for the server.
func WithLogger(l *zap.Logger) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.Logger = l
	}
}

// WithUploadPoolSize sets the upload pool size for the server.
func WithUploadPoolSize(size int) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.UploadPoolSize = size
	}
}

// WithDeletePoolSize sets the delete pool size for the server.
func WithDeletePoolSize(size int) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.DeletePoolSize = size
	}
}

// WithShutdownTimeout sets the shutdown timeout for the server.
func WithShutdownTimeout(timeout time.Duration) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ShutdownTimeout = timeout
	}
}

// defaultServerConfig returns a default server configuration.
func defaultServerConfig() *ServerConfig {
	return &ServerConfig{
		HTTPAddr:        ":9500",
		Logger:          nil,
		UploadPoolSize:  1,
		DeletePoolSize:  1,
		ShutdownTimeout: 10 * time.Second,
	}
}

// Server starts the server with the given options.
func ListenAndServe(options ...ServerOption) {
	cfg := defaultServerConfig()
	for _, opt := range options {
		opt(cfg)
	}

	if cfg.Logger == nil {
		cfg.Logger, _ = zap.NewDevelopment()
	}

	cfg.Logger = cfg.Logger.Named("HTTP")

	defer cfg.Logger.Sync()
	zap.ReplaceGlobals(cfg.Logger)

	// Initialize worker pools
	uploadPool, err := worker.NewPool(cfg.UploadPoolSize)
	if err != nil {
		cfg.Logger.Fatal("Failed to create upload pool", zap.Error(err))
	}
	defer uploadPool.Release()

	deletePool, err := worker.NewPool(cfg.DeletePoolSize)
	if err != nil {
		cfg.Logger.Fatal("Failed to create delete pool", zap.Error(err))
	}
	defer deletePool.Release()

	// Create storage handler
	baseHandler := handlers.NewStorageHandler(uploadPool, deletePool)
	handler := middleware.ChainMiddleware(baseHandler,
		middleware.PathValidationMiddleware,
		middleware.LoggingMiddleware,
	)

	// Address
	if cfg.HTTPAddr == "" {
		val, ok := os.LookupEnv("STORAGE_ADDR")
		if ok {
			cfg.HTTPAddr = val
		}
	}

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	// Setup server with middleware
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: c.Handler(handler),
	}

	// Graceful shutdown channel
	serverErr := make(chan error, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		cfg.Logger.Info("Starting server", zap.String("address", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal
	select {
	case err := <-serverErr:
		cfg.Logger.Fatal("Server failed", zap.Error(err))
	case <-stop:
		cfg.Logger.Info("Received shutdown signal")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		// Initiate graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			cfg.Logger.Error("Server shutdown error", zap.Error(err))
		}

		// Shutdown additional components
		baseHandler.Shutdown()
		cfg.Logger.Info("Server stopped gracefully")
	}
}
