package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/henryhall897/golang-todo-app/gen/mocks/usersmock"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"github.com/henryhall897/golang-todo-app/internal/users/services"
	"github.com/henryhall897/golang-todo-app/internal/users/testutils"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupSuite holds shared test dependencies
type HandlerTestSuite struct {
	mockService *usersmock.ServiceMock
	handler     *Handler
	router      *http.ServeMux
}

// SetupSuite initializes common dependencies but does NOT define routes
func SetupSuite() *HandlerTestSuite {
	logger := zap.NewNop().Sugar() // No-op logger for tests

	mockService := &usersmock.ServiceMock{}

	handler := &Handler{
		service: mockService,
		logger:  logger,
	}

	router := http.NewServeMux() // Router is initialized but not populated

	return &HandlerTestSuite{
		mockService: mockService,
		handler:     handler,
		router:      router,
	}
}

// Test the CreateUserHandler function
func TestCreateUserHandler(t *testing.T) {
	suite := SetupSuite() // Load shared setup

	// Define test-specific routes with middleware
	suite.router.Handle("/users", MethodHandler("POST", func(w http.ResponseWriter, r *http.Request) {
		VerifyCreateUserBody(http.HandlerFunc(suite.handler.CreateUserHandler())).ServeHTTP(w, r)
	}))

	// Generate a sample user
	sampleUser := testutils.GenerateMockUsers(1)[0]

	t.Run("success - user created", func(t *testing.T) {
		// Prepare request payload
		reqBody, err := json.Marshal(map[string]string{
			"name":  sampleUser.Name,
			"email": sampleUser.Email,
		})
		require.NoError(t, err)

		// Mock successful service call
		suite.mockService.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return domain.User{
				ID:        sampleUser.ID,
				Name:      params.Name,
				Email:     params.Email,
				CreatedAt: sampleUser.CreatedAt,
				UpdatedAt: sampleUser.UpdatedAt,
			}, nil
		}

		// Create an HTTP request
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Record the response
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		// Assertions
		require.Equal(t, http.StatusCreated, rr.Code, "handler returned wrong status code")

		var responseBody domain.User
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		assert.Equal(t, sampleUser.Name, responseBody.Name, "handler returned wrong Name")
		assert.Equal(t, sampleUser.Email, responseBody.Email, "handler returned wrong Email")
		assert.Equal(t, sampleUser.ID, responseBody.ID, "handler returned invalid ID")
		assert.False(t, responseBody.CreatedAt.IsZero(), "handler returned zero CreatedAt")
		assert.False(t, responseBody.UpdatedAt.IsZero(), "handler returned zero UpdatedAt")
	})

	t.Run("failure - invalid request body", func(t *testing.T) {
		// Create request with malformed JSON
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{invalid-json"))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code, "expected bad request status")
		assert.JSONEq(t, `{"code": 400, "message": "`+common.MsgInvalidRequestBody+`"}`, rr.Body.String())
	})

	t.Run("failure - missing required fields", func(t *testing.T) {
		// Create request with missing fields
		reqBody, _ := json.Marshal(map[string]string{"name": ""}) // No email
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code, "expected bad request status")
		assert.JSONEq(t, `{"code": 400, "message": "`+common.MsgInvalidInput+`"}`, rr.Body.String())
	})

	t.Run("failure - email already exists", func(t *testing.T) {
		// Mock service returning ErrEmailAlreadyExists
		suite.mockService.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return domain.User{}, services.ErrEmailAlreadyExists
		}

		reqBody, _ := json.Marshal(map[string]string{"name": sampleUser.Name, "email": sampleUser.Email})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code, "expected conflict status")
		assert.JSONEq(t, `{"code": 409, "message": "`+common.MsgEmailAlreadyExists+`"}`, rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock service returning unexpected error
		suite.mockService.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return domain.User{}, common.ErrInternalServerError
		}

		reqBody, _ := json.Marshal(map[string]string{"name": sampleUser.Name, "email": sampleUser.Email})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code, "expected internal server error status")
		assert.JSONEq(t, `{"code": 500, "message": "`+common.MsgInternalServerError+`"}`, rr.Body.String())
	})
}

// Test the GetUserByIDHandler function
func TestGetUserByIDHandler(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	// Define test-specific routes with middleware
	suite.router.Handle("/users/{id}", MethodHandler("GET", VerifyUserID(http.HandlerFunc(suite.handler.GetUserByIDHandler()))))

	// Generate a sample user
	sampleUser := testutils.GenerateMockUsers(1)[0]

	t.Run("success - user found", func(t *testing.T) {
		// Mock service returning the user
		suite.mockService.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			if id == sampleUser.ID {
				return sampleUser, nil
			}
			return domain.User{}, common.ErrNotFound
		}

		// Create an HTTP request with a valid user ID
		req := httptest.NewRequest(http.MethodGet, "/users/"+sampleUser.ID.String(), nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

		var responseBody domain.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		assert.Equal(t, sampleUser.Name, responseBody.Name, "handler returned wrong Name")
		assert.Equal(t, sampleUser.Email, responseBody.Email, "handler returned wrong Email")
		assert.Equal(t, sampleUser.ID, responseBody.ID, "handler returned wrong ID")
	})

	t.Run("failure - user not found", func(t *testing.T) {
		// Mock service returning ErrNotFound
		suite.mockService.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.New().String(), nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code, "expected not found status")
		assert.JSONEq(t, `{"code": 404, "message": "`+common.MsgNotFound+`"}`, rr.Body.String())
	})

	t.Run("failure - missing user ID in path", func(t *testing.T) {
		// Create request missing the user ID
		req := httptest.NewRequest(http.MethodGet, "/users/", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		// Change expected status from 400 to 404
		require.Equal(t, http.StatusNotFound, rr.Code, "expected not found status")
	})

	t.Run("failure - invalid user ID format", func(t *testing.T) {
		// Create request with an invalid UUID format
		req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code, "expected bad request status")
		assert.JSONEq(t, `{"code": 400, "message": "`+common.MsgInvalidInput+`"}`, rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock service returning ErrInternalServerError
		suite.mockService.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, common.ErrInternalServerError
		}

		req := httptest.NewRequest(http.MethodGet, "/users/"+sampleUser.ID.String(), nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code, "expected internal server error status")
		assert.JSONEq(t, `{"code": 500, "message": "`+common.MsgInternalServerError+`"}`, rr.Body.String())
	})
}

// Test the GetUsersHandler function
func TestGetUsersHandler(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	// Define test-specific routes with middleware
	suite.router.Handle("/users", MethodHandler("GET", VerifyGetUsersQuery(suite.handler.GetUsersHandler())))

	// Generate a sample list of users
	sampleUsers := testutils.GenerateMockUsers(3)

	t.Run("success - users retrieved", func(t *testing.T) {
		// Mock service returning users
		suite.mockService.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return sampleUsers, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

		var responseBody []domain.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		assert.Len(t, responseBody, len(sampleUsers), "handler returned wrong number of users")

		for i, user := range responseBody {
			assert.Equal(t, sampleUsers[i].Name, user.Name, "handler returned wrong Name for user %d", i)
			assert.Equal(t, sampleUsers[i].Email, user.Email, "handler returned wrong Email for user %d", i)
			assert.Equal(t, sampleUsers[i].ID, user.ID, "handler returned wrong ID for user %d", i)
		}
	})

	t.Run("failure - no users found", func(t *testing.T) {
		// Mock service returning ErrNotFound
		suite.mockService.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return []domain.User{}, common.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code, "expected not found status")
		assert.JSONEq(t, `{"code": 404, "message": "`+common.MsgNotFound+`"}`, rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock service returning ErrInternalServerError
		suite.mockService.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return []domain.User{}, common.ErrInternalServerError
		}

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code, "expected internal server error status")
		assert.JSONEq(t, `{"code": 500, "message": "`+common.MsgInternalServerError+`"}`, rr.Body.String())
	})
}

// Test the GetUserByEmailHandler function
func TestGetUserByEmailHandler(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	// Define test-specific routes with middleware
	suite.router.Handle("/users", MethodHandler("GET", VerifyGetUsersQuery(suite.handler.GetUserByEmailHandler())))

	// Generate a sample user
	sampleUser := testutils.GenerateMockUsers(1)[0]

	t.Run("success - user found", func(t *testing.T) {
		// Mock service returning the user
		suite.mockService.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			if email == sampleUser.Email {
				return sampleUser, nil
			}
			return domain.User{}, common.ErrNotFound
		}

		// Create an HTTP request with the sample user's email
		req := httptest.NewRequest(http.MethodGet, "/users?email="+sampleUser.Email, nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

		var responseBody domain.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		assert.Equal(t, sampleUser.Name, responseBody.Name, "handler returned wrong Name")
		assert.Equal(t, sampleUser.Email, responseBody.Email, "handler returned wrong Email")
		assert.Equal(t, sampleUser.ID, responseBody.ID, "handler returned wrong ID")
	})

	t.Run("failure - user not found", func(t *testing.T) {
		// Mock service returning ErrNotFound
		suite.mockService.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/users?email=nonexistent@example.com", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code, "expected not found status")
		assert.JSONEq(t, `{"code": 404, "message": "`+common.MsgNotFound+`"}`, rr.Body.String())
	})

	t.Run("failure - missing email parameter", func(t *testing.T) {
		// Create a request without the email parameter
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code, "expected not found status")
		assert.JSONEq(t, `{"code": 404, "message": "`+common.MsgNotFound+`"}`, rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock service returning ErrInternalServerError
		suite.mockService.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, common.ErrInternalServerError
		}

		req := httptest.NewRequest(http.MethodGet, "/users?email="+sampleUser.Email, nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code, "expected internal server error status")
		assert.JSONEq(t, `{"code": 500, "message": "`+common.MsgInternalServerError+`"}`, rr.Body.String())
	})
}

/* TODO fix this test case for new implementation
func TestUpdateUserHandler(t *testing.T) {
	// Generate a sample user
	sampleUser := GenerateMockUsers(1)[0]

	// Prepare mock store
	mockStore := &usersmock.RepositoryMock{
		UpdateUserFunc: func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			if params.ID == sampleUser.ID {
				// Simulate updating the user
				updatedUser := sampleUser
				updatedUser.Name = params.Name
				updatedUser.Email = params.Email
				return updatedUser, nil
			}
			return domain.User{}, common.ErrNotFound
		},
	}

	// Initialize mock logger and handler
	logger := zap.NewNop().Sugar()
	handler := &Handler{
		service: mockStore,
		logger:  logger,
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
		var responseBody domain.User
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
}*/

/* TODO fix this test case for new implementation
func TestDeleteUserHandler(t *testing.T) {
	// Generate a sample user
	sampleUser := GenerateMockUsers(1)[0]

	// Prepare mock store
	mockStore := &usersmock.RepositoryMock{
		DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
			if id == sampleUser.ID {
				return nil // Simulate successful deletion
			}
			return common.ErrNotFound // Simulate user not found
		},
	}

	// Initialize mock logger and handler
	logger := zap.NewNop().Sugar()
	handler := &Handler{
		service: mockStore,
		logger:  logger,
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
} */
