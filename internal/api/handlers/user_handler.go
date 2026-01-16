// Package handlers provides HTTP request handlers
// Complies with CODING_STANDARDS.md: HTTP handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/divyang-garg/sentinel-hub-api/internal/models"
	"github.com/divyang-garg/sentinel-hub-api/internal/services"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /api/v1/users
// Creates a new user account
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req services.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	user, err := h.userService.CreateUser(ctx, &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /api/v1/users/{id}
// Retrieves a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userService.GetUser(ctx, id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /api/v1/users/{id}
// Updates an existing user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req services.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	user, err := h.userService.UpdateUser(ctx, id, &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/users/{id}
// Deletes a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(ctx, id); err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods for consistent responses

func (h *UserHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *UserHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *models.ValidationError:
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "validation_failed",
			"field":   e.Field,
			"message": e.Message,
		})
	case *models.NotFoundError:
		h.respondError(w, http.StatusNotFound, e.Error())
	case *models.AuthenticationError:
		h.respondError(w, http.StatusUnauthorized, e.Message)
	case *models.AuthorizationError:
		h.respondError(w, http.StatusForbidden, e.Message)
	default:
		h.respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
