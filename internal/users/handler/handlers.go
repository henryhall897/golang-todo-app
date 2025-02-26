package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

type Handler struct {
	service domain.Service
	logger  *zap.SugaredLogger
}

// NewUserHandler initializes a new UserHandler instance
func NewUserHandler(service domain.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateUserHandler handles creating a new user
func (h *Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserParams

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteJSONError(w, http.StatusBadRequest, common.MsgInvalidRequestBody)
		return
	}

	// Validate input
	if req.Name == "" || req.Email == "" {
		common.WriteJSONError(w, http.StatusBadRequest, common.MsgMissingFields)
		return
	}

	// Call service layer
	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		if errors.Is(err, common.ErrEmailAlreadyExists) {
			common.WriteJSONError(w, http.StatusConflict, common.MsgEmailAlreadyExists)
		} else {
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
		}
		return
	}

	// Return created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		common.WriteJSONError(w, http.StatusInternalServerError, common.MsgFailedEncoding)
	}
}

func (h *Handler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract userID from context
	userIDStr, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "User ID not found", http.StatusBadRequest)
		return
	}

	// Convert userID to UUID and fetch the user
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("Failed to get user by ID", "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return the user as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetUserByEmailHandler handles retrieving a user by email
func (h *Handler) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract email from query parameters
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	// Call store to get the user by email
	user, err := h.service.GetUserByEmail(r.Context(), email)
	if errors.Is(err, common.ErrNotFound) {
		h.logger.Errorw("User not found", "email", email)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Errorw("Failed to get user by email", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the user as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListUsersHandler handles retrieving all users
func (h *Handler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Call store to list users using the request's context
	userList, err := h.service.ListUsers(r.Context(), domain.ListUsersParams{})
	if err != nil {
		h.logger.Errorw("Failed to list users", "error", err, "path", r.URL.Path, "method", r.Method)
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	// Handle empty response case
	if len(userList) == 0 {
		userList = []domain.User{}
	}

	// Set response headers and write status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(userList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateUserHandler handles updating a user's information
func (h *Handler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userIDStr, ok := r.Context().Value(userIDKey).(string)
	if !ok || userIDStr == "" {
		http.Error(w, "User ID not found", http.StatusBadRequest)
		return
	}

	// Parse the user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Parse the request body
	var payload struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Errorw("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure at least one field is provided
	if payload.Name == nil || payload.Email == nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Prepare update parameters
	updateUserParams := domain.UpdateUserParams{
		ID:    userID,
		Name:  *payload.Name,
		Email: *payload.Email,
	}

	// Call the store to update the user
	updatedUser, err := h.service.UpdateUser(r.Context(), updateUserParams)
	if errors.Is(err, common.ErrNotFound) {
		h.logger.Errorw("User not found", "userID", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Errorw("Failed to update user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the updated user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		h.logger.Errorw("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// DeleteUserHandler handles deleting a user by ID
func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the context
	userIDStr, ok := r.Context().Value(userIDKey).(string)
	if !ok || userIDStr == "" {
		http.Error(w, "User ID not found", http.StatusBadRequest)
		return
	}

	// Parse the user ID
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Call store to delete the user
	err = h.service.DeleteUser(r.Context(), id)
	if errors.Is(err, common.ErrNotFound) {
		h.logger.Errorw("User not found", "userID", id)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Errorw("Failed to delete user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return a success response (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}
