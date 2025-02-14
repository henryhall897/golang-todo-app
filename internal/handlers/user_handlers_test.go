package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// GenerateMockUsers creates a specified number of mock users with unique emails.
func GenerateMockUsers(count int) []users.User {
	now := time.Now()
	userList := make([]users.User, count)
	for i := 0; i < count; i++ {
		userList[i] = users.User{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("John %d Doe", i+1),
			Email:     fmt.Sprintf("johndoe%d@example.com", i+1),
			CreatedAt: &now,
			UpdatedAt: &now,
		}
	}
	return userList
}

func TestCreateUserHandler(t *testing.T) {
	// Generate a sample user
	sampleUser := GenerateMockUsers(1)[0]

	// Prepare mock store
	mockStore := &MockStore{
		CreateUserFunc: func(ctx context.Context, params users.CreateUserParams) (users.User, error) {
			return users.User{
				ID:        sampleUser.ID,
				Name:      params.Name,
				Email:     params.Email,
				CreatedAt: sampleUser.CreatedAt,
				UpdatedAt: sampleUser.UpdatedAt,
			}, nil
		},
	}

	// Initialize mock logger and handler
	mockLogger := &MockLogger{}
	handler := &UserHandler{
		Store:  mockStore,
		Logger: mockLogger,
	}

	// Setup router with middleware and route
	router := http.NewServeMux()
	router.Handle("/users", MethodHandler("POST", handler.CreateUserHandler))

	// Prepare request payload using the sample user's data
	reqBody, err := json.Marshal(map[string]string{
		"name":  sampleUser.Name,
		"email": sampleUser.Email,
	})
	require.NoError(t, err)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token") // Simulate valid token for auth

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Send the request through the router (which includes middleware)
	router.ServeHTTP(rr, req)

	// Assert the status code
	require.Equal(t, http.StatusCreated, rr.Code, "handler returned wrong status code")

	// Decode and verify the response body
	var responseBody users.User
	err = json.NewDecoder(rr.Body).Decode(&responseBody)
	require.NoError(t, err, "failed to decode response body")

	// Check the returned fields
	require.Equal(t, sampleUser.Name, responseBody.Name, "handler returned wrong Name")
	require.Equal(t, sampleUser.Email, responseBody.Email, "handler returned wrong Email")
	require.Equal(t, sampleUser.ID, responseBody.ID, "handler returned invalid ID")
	require.False(t, responseBody.CreatedAt.IsZero(), "handler returned zero CreatedAt")
	require.False(t, responseBody.UpdatedAt.IsZero(), "handler returned zero UpdatedAt")
}

func TestGetUserByIDHandler(t *testing.T) {
	// Generate a sample user
	sampleUser := GenerateMockUsers(1)[0]

	// Prepare mock store
	mockStore := &MockStore{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (users.User, error) {
			if id == sampleUser.ID {
				return sampleUser, nil
			}
			return users.User{}, &common.UserIDNotFoundError{UserID: id}
		},
	}

	// Initialize mock logger and handler
	mockLogger := &MockLogger{}
	handler := &UserHandler{
		Store:  mockStore,
		Logger: mockLogger,
	}

	// Define dynamic route handlers
	dynamicHandlers := map[string]http.HandlerFunc{
		"GET": handler.GetUserByIDHandler,
	}

	// Setup the router with dynamic route handling
	router := http.NewServeMux()
	router.Handle("/users/", DynamicRouteHandler(dynamicHandlers))

	// Sub-tests
	t.Run("Successful user retrieval", func(t *testing.T) {
		// Create a new HTTP request with the sample user's ID
		req := httptest.NewRequest(http.MethodGet, "/users/"+sampleUser.ID.String(), nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router (which includes dynamic route handling)
		router.ServeHTTP(rr, req)

		// Log or assert the full response body if the status code is not what you expect
		if rr.Code != http.StatusOK {
			t.Logf("Response body: %s", rr.Body.String())
		}

		// Assert the status code
		require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

		// Decode and verify the response body
		var responseBody users.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		// Check the returned fields
		require.Equal(t, sampleUser.Name, responseBody.Name, "handler returned wrong Name")
		require.Equal(t, sampleUser.Email, responseBody.Email, "handler returned wrong Email")
		require.Equal(t, sampleUser.ID, responseBody.ID, "handler returned wrong ID")
	})

	t.Run("User not found", func(t *testing.T) {
		// Create a new HTTP request with a non-existent user ID
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/users/"+nonExistentID.String(), nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusNotFound, rr.Code, "handler returned wrong status code")
	})

	t.Run("Invalid user ID format", func(t *testing.T) {
		// Create a new HTTP request with an invalid user ID
		req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	})
}

func TestListUsersHandler(t *testing.T) {
	// Prepare mock store
	mockStore := &MockStore{
		ListUsersFunc: func(ctx context.Context, params users.ListUsersParams) ([]users.User, error) {
			return GenerateMockUsers(3), nil
		},
	}

	// Initialize mock logger and handler
	mockLogger := &MockLogger{}
	handler := &UserHandler{
		Store:  mockStore,
		Logger: mockLogger,
	}

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, "/users", nil)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler's method
	handler.ListUsersHandler(rr, req)

	// Assert the status code
	require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	// Decode and verify the response body
	var responseBody []users.User
	err := json.NewDecoder(rr.Body).Decode(&responseBody)
	require.NoError(t, err, "failed to decode response body")

	// Verify the number of users returned
	require.Len(t, responseBody, 3, "handler returned wrong number of users")

	// Verify the email addresses of the users
	for i, user := range responseBody {
		expectedEmail := fmt.Sprintf("johndoe%d@example.com", i+1)
		require.Equal(t, expectedEmail, user.Email, "handler returned wrong email for user %d", i+1)
	}
}

func TestGetUserByEmailHandler(t *testing.T) {
	// Generate mock users
	mockUsers := GenerateMockUsers(3)
	sampleUser := mockUsers[0] // We'll use the first user for testing

	// Prepare mock store
	mockStore := &MockStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (users.User, error) {
			// Search for the user by email in mock data
			for _, user := range mockUsers {
				if user.Email == email {
					return user, nil
				}
			}
			return users.User{}, common.ErrNotFound
		},
	}

	// Initialize mock logger and handler
	mockLogger := &MockLogger{}
	handler := &UserHandler{
		Store:  mockStore,
		Logger: mockLogger,
	}

	// Setup the router and register the route
	router := http.NewServeMux()
	router.Handle("/users/email", http.HandlerFunc(handler.GetUserByEmailHandler))

	// Sub-tests
	t.Run("Successful user retrieval", func(t *testing.T) {
		// Create a request with the sample user's email
		req := httptest.NewRequest(http.MethodGet, "/users/email?email="+sampleUser.Email, nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

		// Decode and verify the response body
		var responseBody users.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		// Check the returned fields
		require.Equal(t, sampleUser.Name, responseBody.Name, "handler returned wrong Name")
		require.Equal(t, sampleUser.Email, responseBody.Email, "handler returned wrong Email")
		require.Equal(t, sampleUser.ID, responseBody.ID, "handler returned wrong ID")
	})

	t.Run("User not found", func(t *testing.T) {
		// Create a request with a non-existent email
		req := httptest.NewRequest(http.MethodGet, "/users/email?email=nonexistent@example.com", nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusNotFound, rr.Code, "handler returned wrong status code")
	})

	t.Run("Missing email parameter", func(t *testing.T) {
		// Create a request without the email parameter
		req := httptest.NewRequest(http.MethodGet, "/users/email", nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	})
}

func TestUpdateUserHandler(t *testing.T) {
	// Generate a sample user
	sampleUser := GenerateMockUsers(1)[0]

	// Prepare mock store
	mockStore := &MockStore{
		UpdateUserFunc: func(ctx context.Context, params users.UpdateUserParams) (users.User, error) {
			if params.ID == sampleUser.ID {
				// Simulate updating the user
				updatedUser := sampleUser
				updatedUser.Name = *params.Name
				updatedUser.Email = *params.Email
				return updatedUser, nil
			}
			return users.User{}, common.ErrNotFound
		},
	}

	// Initialize mock logger and handler
	mockLogger := &MockLogger{}
	handler := &UserHandler{
		Store:  mockStore,
		Logger: mockLogger,
	}

	// Define dynamic route handlers
	dynamicHandlers := map[string]http.HandlerFunc{
		"PUT": handler.UpdateUserHandler,
	}

	// Setup the router with dynamic route handling
	router := http.NewServeMux()
	router.Handle("/users/", DynamicRouteHandler(dynamicHandlers))

	// Sub-tests

	t.Run("Successful user update", func(t *testing.T) {
		// New data to update
		newName := "Updated John Doe"
		newEmail := "updated.johndoe@example.com"

		// Prepare request payload
		reqBody, err := json.Marshal(map[string]string{
			"name":  newName,
			"email": newEmail,
		})
		require.NoError(t, err)

		// Create a new HTTP request with the sample user's ID and update payload
		req := httptest.NewRequest(http.MethodPut, "/users/"+sampleUser.ID.String(), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

		// Decode and verify the response body
		var responseBody users.User
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		// Check the updated fields
		require.Equal(t, newName, responseBody.Name, "handler returned wrong updated Name")
		require.Equal(t, newEmail, responseBody.Email, "handler returned wrong updated Email")
		require.Equal(t, sampleUser.ID, responseBody.ID, "handler returned wrong ID")
	})

	t.Run("User not found", func(t *testing.T) {
		// Arrange: Create a non-existent user ID and a valid request body
		nonExistentID := uuid.New()
		reqBody, err := json.Marshal(map[string]string{
			"name":  "Test Name",
			"email": "test@example.com",
		})
		require.NoError(t, err)

		// Create a new HTTP request with the non-existent user ID and valid request body
		req := httptest.NewRequest(http.MethodPut, "/users/"+nonExistentID.String(), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Act: Send the request through the router
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Assert: Check that the response status code is 404 Not Found
		require.Equal(t, http.StatusNotFound, rr.Code, "handler returned wrong status code")
	})

	t.Run("Invalid user ID format", func(t *testing.T) {
		// Create a new HTTP request with an invalid user ID
		req := httptest.NewRequest(http.MethodPut, "/users/invalid-uuid", nil)
		req.Header.Set("Content-Type", "application/json")

		// Send the request through the router
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	})

	t.Run("Invalid request body", func(t *testing.T) {
		// Create a new HTTP request with invalid JSON body
		req := httptest.NewRequest(http.MethodPut, "/users/"+sampleUser.ID.String(), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		// Send the request through the router
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	})
}

func TestDeleteUserHandler(t *testing.T) {
	// Generate a sample user
	sampleUser := GenerateMockUsers(1)[0]

	// Prepare mock store
	mockStore := &MockStore{
		DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
			if id == sampleUser.ID {
				return nil // Simulate successful deletion
			}
			return common.ErrNotFound // Simulate user not found
		},
	}

	// Initialize mock logger and handler
	mockLogger := &MockLogger{}
	handler := &UserHandler{
		Store:  mockStore,
		Logger: mockLogger,
	}

	// Define dynamic route handlers
	dynamicHandlers := map[string]http.HandlerFunc{
		"DELETE": handler.DeleteUserHandler,
	}

	// Setup the router with dynamic route handling
	router := http.NewServeMux()
	router.Handle("/users/", DynamicRouteHandler(dynamicHandlers))

	// Sub-tests

	t.Run("Successful user deletion", func(t *testing.T) {
		// Create a new HTTP request with the sample user's ID
		req := httptest.NewRequest(http.MethodDelete, "/users/"+sampleUser.ID.String(), nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusNoContent, rr.Code, "handler returned wrong status code")
	})

	t.Run("User not found", func(t *testing.T) {
		// Create a new HTTP request with a non-existent user ID
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/users/"+nonExistentID.String(), nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusNotFound, rr.Code, "handler returned wrong status code")
	})

	t.Run("Invalid user ID format", func(t *testing.T) {
		// Create a new HTTP request with an invalid user ID
		req := httptest.NewRequest(http.MethodDelete, "/users/invalid-uuid", nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	})

	t.Run("Internal server error", func(t *testing.T) {
		// Set up mock to simulate an internal error
		mockStore.DeleteUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return fmt.Errorf("database error")
		}

		// Create a new HTTP request with the sample user's ID
		req := httptest.NewRequest(http.MethodDelete, "/users/"+sampleUser.ID.String(), nil)

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Send the request through the router
		router.ServeHTTP(rr, req)

		// Assert the status code
		require.Equal(t, http.StatusInternalServerError, rr.Code, "handler returned wrong status code")
	})
}
