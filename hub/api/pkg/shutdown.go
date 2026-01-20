// Package pkg provides shared utilities
package pkg

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var isShuttingDown atomic.Bool

// SetShuttingDown sets the shutdown state
func SetShuttingDown(value bool) {
	isShuttingDown.Store(value)
}

// IsShuttingDown checks if the server is shutting down
func IsShuttingDown() bool {
	return isShuttingDown.Load()
}

// CleanupFunc is a function type for cleanup operations
type CleanupFunc func(ctx context.Context)

// GracefulShutdown performs graceful shutdown of the HTTP server
func GracefulShutdown(server *http.Server, cleanup CleanupFunc) {
	log.Println("Initiating graceful shutdown...")
	SetShuttingDown(true)

	// Phase 1: Stop accepting new requests (30s timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Phase 2: Complete in-flight requests
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Forced shutdown: %v", err)
	}

	// Phase 3: Cleanup resources (10s timeout)
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()

	if cleanup != nil {
		cleanup(cleanupCtx)
	}

	log.Println("Shutdown complete")
}
