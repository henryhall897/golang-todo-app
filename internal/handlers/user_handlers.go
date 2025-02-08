package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"golang-todo-app/internal/core/common"
	"golang-todo-app/internal/core/logging"
	"golang-todo-app/internal/users"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	Store  users.Store
	Logger logging.Logger
}

// CreateUserHandler handles creating a new user
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call store to create the user
	user, err := h.Store.CreateUser(context.Background(), req.Name, req.Email)
	if err != nil {
		h.Logger.Errorw("Failed to create user", "error", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Return the created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUserByIDHandler handles retrieving a user by ID
func (h *UserHandler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path parameters
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Call store to get the user
	user, err := h.Store.GetUserByID(context.Background(), id)
	if err != nil {
		h.Logger.Errorw("Failed to get user", "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return the user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetUserByEmailHandler handles retrieving a user by email
func (h *UserHandler) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract email from query parameters
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Call store to get the user by email
	user, err := h.Store.GetUserByEmail(context.Background(), email)
	if errors.Is(err, common.ErrNotFound) {
		h.Logger.Errorw("User not found", "email", email)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.Logger.Errorw("Failed to get user by email", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ListUsersHandler handles retrieving all users
func (h *UserHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Call store to list users
	users, err := h.Store.ListUsers(context.Background(), users.ListUsersParams{})
	if err != nil {
		h.Logger.Errorw("Failed to list users", "error", err)
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	// Return the list of users
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// UpdateUserHandler handles updating a user's information
func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the path parameters
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body to extract the name and email
	var payload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input fields
	if payload.Name == "" || payload.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	// Call store to update the user
	updatedUser, err := h.Store.UpdateUser(context.Background(), id, payload.Name, payload.Email)
	if errors.Is(err, common.ErrNotFound) {
		h.Logger.Errorw("User not found", "id", id)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.Logger.Errorw("Failed to update user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the updated user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

// DeleteUserHandler handles deleting a user by ID
func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the path parameters
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Call store to delete the user
	err = h.Store.DeleteUser(context.Background(), id)
	if errors.Is(err, common.ErrNotFound) {
		h.Logger.Errorw("User not found", "id", id)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.Logger.Errorw("Failed to delete user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusNoContent)
}
