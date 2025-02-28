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
func (h *Handler) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract validated request from context
		req, ok := r.Context().Value(validatedUserKey).(domain.CreateUserParams)
		if !ok {
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
			return
		}

		// Call service layer
		user, err := h.service.CreateUser(r.Context(), req)
		if err != nil {
			if errors.Is(err, services.ErrEmailAlreadyExists) {
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
}

func (h *Handler) GetUserByIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract validated user ID from context
		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			h.logger.Errorw("GetUserByID failed: user ID missing in request context")
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
			return
		}

		// Call the service layer
		user, err := h.service.GetUserByID(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, common.ErrNotFound):
				common.WriteJSONError(w, http.StatusNotFound, common.MsgNotFound)
				return
			default:
				common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
				return
			}
		}

		// Return the user as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			h.logger.Errorw("GetUserByID failed: failed to encode response", "user_id", userID, "error", err)
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgFailedEncoding)
		}
	}
}

// GetUserByEmailHandler handles retrieving a user by email
func (h *Handler) GetUserByEmailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve query parameters from context
		queryParams, ok := r.Context().Value(queryParamsKey).(domain.QueryParams)
		if !ok {
			h.logger.Errorw("GetUserByEmail failed: missing query parameters in context")
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
			return
		}

		// Call the service layer
		user, err := h.service.GetUserByEmail(r.Context(), queryParams.Email)
		if err != nil {
			switch {
			case errors.Is(err, common.ErrNotFound):
				common.WriteJSONError(w, http.StatusNotFound, common.MsgNotFound)
				return
			default:
				common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
				return
			}
		}

		// Return the user as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			h.logger.Errorw("GetUserByEmail failed: failed to encode response", "email", queryParams.Email, "error", err)
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgFailedEncoding)
		}
	}
}

func (h *Handler) GetUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve query parameters from context
		queryParams, ok := r.Context().Value(queryParamsKey).(domain.QueryParams)
		if !ok {
			h.logger.Errorw("GetUsersHandler failed: missing query parameters in context")
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
			return
		}

		// If email is provided, redirect to GetUserByEmailHandler
		if queryParams.QueryType == domain.QueryTypeEmail {
			h.GetUserByEmailHandler().ServeHTTP(w, r)
			return
		}

		// Construct user query params for listing users
		getUsersParams := domain.GetUsersParams{
			Limit:  queryParams.Limit,
			Offset: queryParams.Offset,
		}

		// Call service layer
		users, err := h.service.GetUsers(r.Context(), getUsersParams)
		if err != nil {
			if errors.Is(err, common.ErrNotFound) {
				common.WriteJSONError(w, http.StatusNotFound, common.MsgNotFound)
				return
			}
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgInternalServerError)
			return
		}

		// Ensure empty array instead of nil
		if len(users) == 0 {
			users = []domain.User{}
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(users); err != nil {
			h.logger.Errorw("GetUsersHandler failed: failed to encode response", "params", getUsersParams, "error", err)
			common.WriteJSONError(w, http.StatusInternalServerError, common.MsgFailedEncoding)
		}
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
