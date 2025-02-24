package backend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/veloxpack/storage/pkg/backend/server"
	"github.com/veloxpack/storage/pkg/storage"
	"github.com/veloxpack/storage/pkg/storage/provider"
)

// StorageBackend manages storage backend and server configuration
type StorageBackend struct {
	provider provider.Storage
}

// Client returns the configured storage provider
func (c *StorageBackend) Client() provider.Storage {
	return c.provider
}

// Server creates a new HTTP server with the provided options and the storage backend's provider.
// It returns the configured HTTP server or an error if the server creation fails.
//
// Parameters:
//
//	opts - A variadic list of server options to configure the HTTP server.
//
// Returns:
//
//	*http.Server - The configured HTTP server.
//	error - An error if the server creation fails.
func (c *StorageBackend) Server(opts ...server.ServerOption) (*http.Server, error) {
	// Prepend client-specific configuration to user options
	serverOpts := append(
		[]server.ServerOption{server.WithBackend(c.provider)},
		opts...,
	)

	httpServer, err := server.NewServer(serverOpts...)
	if err != nil {
		return nil, err
	}

	return httpServer, nil
}

// ListenAndServe starts an HTTP server with the provided options and listens on a random port.
// It returns the server's address, a shutdown function, and an error if the server creation fails.
//
// Parameters:
//
//	opts - A variadic list of server options to configure the HTTP server.
//
// Returns:
//
//	string - The address of the server in the format "http://localhost:<port>".
//	func(context.Context) error - A function to gracefully shut down the server.
//	error - An error if the server creation fails.
func (c *StorageBackend) ListenAndServe(opts ...server.ServerOption) (string, func(context.Context) error, error) {
	// Create listener on random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create listener: %w", err)
	}

	// Create HTTP server with options
	httpServer, err := c.Server(opts...)
	if err != nil {
		listener.Close()
		return "", nil, fmt.Errorf("failed to create server: %w", err)
	}

	// Get actual listening port
	addr := listener.Addr().(*net.TCPAddr)
	host := fmt.Sprintf("http://localhost:%d", addr.Port)

	// Start server in goroutine
	go func() {
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Return shutdown function that gracefully stops the server
	shutdown := func(ctx context.Context) error {
		return httpServer.Shutdown(ctx)
	}

	return host, shutdown, nil
}

// NewStorageBackend creates a new storage backend with environment variables as defaults
func NewStorageBackend(opts ...storage.StorageOption) *StorageBackend {
	// Create default options from environment variables
	defaultOpts := []storage.StorageOption{
		storage.WithDriver(os.Getenv("STORAGE_DRIVER")),
		storage.WithOutputLocation(os.Getenv("STORAGE_OUTPUT_LOCATION")),
	}

	// Prepend default options so user options can override them
	mergedOpts := append(defaultOpts, opts...)

	return &StorageBackend{
		provider: storage.NewStorage(mergedOpts...),
	}
}
