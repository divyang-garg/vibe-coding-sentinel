// Sentinel Hub API Server - Entry Point
// Complies with CODING_STANDARDS.md: Entry Points max 50 lines
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"sentinel-hub-api/config"
	"sentinel-hub-api/handlers"
	"sentinel-hub-api/pkg"
	"sentinel-hub-api/pkg/metrics"
	"sentinel-hub-api/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Initialize database connection
	var db *sql.DB
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err = sql.Open("postgres", dsn)
	} else {
		db, err = sql.Open("postgres", cfg.GetDSN())
	}
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()
	
	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	
	m := metrics.NewMetrics("sentinel_hub_api")
	go metrics.StartSystemMetricsCollection(m)
	deps := handlers.NewDependencies(db)
	r := router.NewRouter(deps, m)
	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	go func() {
		log.Printf("Server starting on %s", cfg.GetServerAddr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	pkg.GracefulShutdown(server, deps.Cleanup)
}
