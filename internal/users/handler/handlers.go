package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"github.com/henryhall897/golang-todo-app/internal/users/services"
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
	// Extract and decode request body
	var params domain.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		h.logger.Warnw("CreateUser failed: invalid request body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Validate input fields
	if params.Name == "" || params.Email == "" {
		h.logger.Warnw("CreateUser failed: missing required fields", "name", params.Name, "email", params.Email)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Call service layer
	user, err := h.service.CreateUser(r.Context(), params)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		} else {
			h.logger.Errorw("CreateUser failed: internal server error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Return created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Errorw("CreateUser failed: failed to encode response", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// GetUserByIDHandler handles retrieving a user by ID
func (h *Handler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract validated user ID from context
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		h.logger.Errorw("GetUserByID failed: user ID missing in request context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Call the service layer
	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		h.logger.Errorw("GetUserByID failed: internal server error", "user_id", userID, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Return the user as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Errorw("GetUserByID failed: failed to encode response", "user_id", userID, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// GetUserByEmailHandler handles retrieving a user by email
func (h *Handler) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract validated query parameters from context
	queryParams, ok := r.Context().Value(queryParamsKey).(domain.GetQueryParams)
	if !ok {
		h.logger.Errorw("GetUserByEmail failed: missing query parameters in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Call the service layer
	user, err := h.service.GetUserByEmail(r.Context(), queryParams.Email)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Return the user as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Errorw("GetUserByEmail failed: failed to encode response", "email", queryParams.Email, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// GetUsersHandler handles retrieving a list of users or a user by email
func (h *Handler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Extract validated query parameters from context
	queryParams, ok := r.Context().Value(queryParamsKey).(domain.GetQueryParams)
	if !ok {
		h.logger.Errorw("GetUsersHandler failed: missing query parameters in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// If email is provided, delegate request to GetUserByEmailHandler
	if queryParams.QueryType == domain.QueryTypeEmail {
		h.GetUserByEmailHandler(w, r)
		return
	}

	// Construct user query params for listing users
	getUsersParams := domain.GetUsersParams{
		Limit:  queryParams.Limit,
		Offset: queryParams.Offset,
	}

	// Call the service layer
	users, err := h.service.GetUsers(r.Context(), getUsersParams)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		h.logger.Errorw("GetUsersHandler failed: internal server error", "params", getUsersParams, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Ensure an empty array instead of nil
	if len(users) == 0 {
		users = []domain.User{}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		h.logger.Errorw("GetUsersHandler failed: failed to encode response", "params", getUsersParams, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// UpdateUserHandler handles updating a user's information
func (h *Handler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract validated user ID from context
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.logger.Errorw("UpdateUserHandler failed: missing or invalid user ID in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Parse the request body
	var payload struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Errorw("UpdateUserHandler failed: invalid request body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Ensure at least one field is provided
	if payload.Name == nil && payload.Email == nil {
		h.logger.Warnw("UpdateUserHandler failed: no fields provided for update")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Prepare update parameters
	updateUserParams := domain.UpdateUserParams{
		ID:    userID,
		Name:  *payload.Name,
		Email: *payload.Email,
	}

	// Call the service layer
	updatedUser, err := h.service.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		h.logger.Errorw("UpdateUserHandler failed: internal server error", "user_id", userID, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Return the updated user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		h.logger.Errorw("UpdateUserHandler failed: failed to encode response", "user_id", userID, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// DeleteUserHandler handles deleting a user by ID
func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract validated user ID from context
	id, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok || id == uuid.Nil {
		h.logger.Errorw("DeleteUserHandler failed: missing or invalid user ID in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Call service layer to delete the user
	err := h.service.DeleteUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Return success response (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}
