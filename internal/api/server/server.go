// Package server provides HTTP server setup and management
// Complies with CODING_STANDARDS.md: Server setup max 200 lines
package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/divyang-garg/sentinel-hub-api/internal/api/handlers"
	"github.com/divyang-garg/sentinel-hub-api/internal/api/middleware"
	"github.com/divyang-garg/sentinel-hub-api/internal/config"
	"github.com/divyang-garg/sentinel-hub-api/internal/repository"
	"github.com/divyang-garg/sentinel-hub-api/internal/services"
)

// Server represents the HTTP server
type Server struct {
	cfg        *config.Config
	db         *sql.DB
	router     *chi.Mux
	httpServer *http.Server
}

// NewServer creates a new HTTP server with all dependencies
func NewServer(cfg *config.Config, db *sql.DB) *Server {
	// Create service and repository dependencies
	userRepo := repository.NewPostgresUserRepository(db)
	passwordHasher := services.NewBcryptPasswordHasher(cfg.Security.BcryptCost)
	userService := services.NewPostgresUserService(userRepo, passwordHasher)

	// Create handlers
	userHandler := handlers.NewUserHandler(userService)

	// Create router with middleware
	r := setupRouter(cfg, userHandler)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		cfg:        cfg,
		db:         db,
		router:     r,
		httpServer: httpServer,
	}
}

// setupRouter creates and configures the router with middleware and routes
func setupRouter(cfg *config.Config, userHandler *handlers.UserHandler) *chi.Mux {
	r := chi.NewRouter()

	// Core middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Security.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Custom middleware
	rateLimiter := middleware.NewRateLimiter(
		cfg.Security.RateLimitRequests,
		cfg.Security.RateLimitWindow,
	)
	r.Use(rateLimiter.RateLimit)

	// Health check (no auth required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// API routes with authentication
	r.Route("/api/v1", func(r chi.Router) {
		// Auth middleware for protected routes
		authMiddleware := middleware.NewAuthMiddleware(cfg.Security.JWTSecret)
		r.Use(authMiddleware.Authenticate)

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.CreateUser)
			r.Get("/{id}", userHandler.GetUser)
			r.Put("/{id}", userHandler.UpdateUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})
	})

	return r
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
