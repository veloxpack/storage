package main

import (
	"context"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mediaprodcast/commons/discovery"
	"github.com/mediaprodcast/commons/env"
	"github.com/mediaprodcast/commons/tracer"
	"go.uber.org/zap"
)

var httpAddr = env.GetString("HTTP_ADDR", ":9500")

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	// Set up global tracing for the application
	if err := tracer.SetGlobalTracer(context.TODO(), discovery.ProbeSvsName); err != nil {
		logger.Fatal("could set global tracer", zap.Error(err))
	}

	ctx := context.Background()

	// Register the service instance in Consul
	instanceID, registry, err := discovery.Register(ctx, discovery.StorageSvsName, httpAddr)
	if err != nil {
		panic(err)
	}

	// Ensure service deregistration on shutdown
	defer registry.Deregister(ctx, instanceID, discovery.StorageSvsName)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Info("Starting storage service", zap.String("port", httpAddr))

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		logger.Fatal("Failed to start storage service")
		panic(err)
	}
}
