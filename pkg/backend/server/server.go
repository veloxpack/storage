package server

import (
	"net/http"

	"github.com/mediaprodcast/storage/pkg/backend/server/handlers"
	"github.com/mediaprodcast/storage/pkg/backend/server/middleware"
	"github.com/mediaprodcast/storage/pkg/backend/server/worker"
	"github.com/mediaprodcast/storage/pkg/storage"
	"github.com/mediaprodcast/storage/pkg/storage/provider"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// ServerOption defines a functional option for configuring the server.
type ServerOption func(*ServerConfig)

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	HTTPAddr       string
	Logger         *zap.Logger
	UploadPoolSize int
	DeletePoolSize int
	backend        provider.Storage
}

// WithHTTPAddr sets the HTTP address for the server.
func WithHTTPAddr(addr string) ServerOption {
	return func(cfg *ServerConfig) {
		if addr != "" {
			cfg.HTTPAddr = addr
		}
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

// WithBackend sets the storage backend for the server.
func WithBackend(backend provider.Storage) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.backend = backend
	}
}

// defaultServerConfig returns a default server configuration.
func defaultServerConfig() *ServerConfig {
	return &ServerConfig{
		UploadPoolSize: 1,
		DeletePoolSize: 1,
		HTTPAddr:       ":9500",
		Logger:         zap.NewNop(),
		backend:        storage.NewStorage(),
	}
}

// Server starts the server with the given options.
func NewServer(options ...ServerOption) (*http.Server, error) {
	cfg := defaultServerConfig()
	for _, opt := range options {
		opt(cfg)
	}

	cfg.Logger = cfg.Logger.Named("HTTP")
	zap.ReplaceGlobals(cfg.Logger)

	// Initialize worker pools
	uploadPool, err := worker.NewPool(cfg.UploadPoolSize)
	if err != nil {
		cfg.Logger.Fatal("Failed to create upload pool", zap.Error(err))
		return nil, err
	}

	defer uploadPool.Release()

	deletePool, err := worker.NewPool(cfg.DeletePoolSize)
	if err != nil {
		cfg.Logger.Fatal("Failed to create delete pool", zap.Error(err))
		return nil, err
	}
	defer deletePool.Release()

	// Create storage handler
	baseHandler := handlers.NewStorageHandler(cfg.backend, uploadPool, deletePool)
	handler := middleware.ChainMiddleware(baseHandler,
		middleware.PathValidationMiddleware,
		middleware.LoggingMiddleware,
	)

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	// Setup server with middleware
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: c.Handler(handler),
	}

	return server, nil
}
