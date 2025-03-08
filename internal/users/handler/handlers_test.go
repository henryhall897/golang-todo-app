package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	suite := SetupSuite()
	suite.router.Handle("/users", MethodHandler("POST", suite.handler.CreateUserHandler))

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
			return sampleUser, nil
		}

		// Create an HTTP request
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Record the response
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code, "handler returned wrong status code")

		var responseBody domain.User
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err, "failed to decode response body")

		// Compare only the relevant fields
		assert.Equal(t, sampleUser.ID, responseBody.ID, "handler returned incorrect ID")
		assert.Equal(t, sampleUser.Name, responseBody.Name, "handler returned incorrect Name")
		assert.Equal(t, sampleUser.Email, responseBody.Email, "handler returned incorrect Email")

		// Instead of direct comparison, check if timestamps are reasonably close
		assert.WithinDuration(t, *sampleUser.CreatedAt, *responseBody.CreatedAt, time.Second, "handler returned incorrect CreatedAt")
		assert.WithinDuration(t, *sampleUser.UpdatedAt, *responseBody.UpdatedAt, time.Second, "handler returned incorrect UpdatedAt")
	})

	t.Run("failure - invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{invalid-json"))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusBadRequest)+"\n", rr.Body.String())
	})

	t.Run("failure - email already exists", func(t *testing.T) {
		suite.mockService.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return domain.User{}, services.ErrEmailAlreadyExists
		}

		reqBody, _ := json.Marshal(map[string]string{"name": sampleUser.Name, "email": sampleUser.Email})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusConflict)+"\n", rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		suite.mockService.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return domain.User{}, common.ErrInternalServerError
		}

		reqBody, _ := json.Marshal(map[string]string{"name": sampleUser.Name, "email": sampleUser.Email})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rr.Body.String())
	})
}

// TestGetUserByIDHandler tests retrieving a user by ID
func TestGetUserByIDHandler(t *testing.T) {
	suite := SetupSuite()
	suite.router.Handle("/users/", MethodHandler("GET", VerifyUserID(http.HandlerFunc(suite.handler.GetUserByIDHandler))))

	sampleUser := testutils.GenerateMockUsers(1)[0]

	t.Run("success - user found", func(t *testing.T) {
		suite.mockService.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return sampleUser, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/users/"+sampleUser.ID.String(), nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var responseBody domain.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err)

		// Compare only the relevant fields
		assert.Equal(t, sampleUser.ID, responseBody.ID, "handler returned incorrect ID")
		assert.Equal(t, sampleUser.Name, responseBody.Name, "handler returned incorrect Name")
		assert.Equal(t, sampleUser.Email, responseBody.Email, "handler returned incorrect Email")

		// Allow slight differences in timestamps
		assert.WithinDuration(t, *sampleUser.CreatedAt, *responseBody.CreatedAt, time.Second, "handler returned incorrect CreatedAt")
		assert.WithinDuration(t, *sampleUser.UpdatedAt, *responseBody.UpdatedAt, time.Second, "handler returned incorrect UpdatedAt")
	})

	t.Run("failure - user not found", func(t *testing.T) {
		suite.mockService.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.New().String(), nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", rr.Body.String())
	})

	t.Run("failure - invalid user ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusBadRequest)+"\n", rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		suite.mockService.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, common.ErrInternalServerError
		}

		req := httptest.NewRequest(http.MethodGet, "/users/"+sampleUser.ID.String(), nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rr.Body.String())
	})
}

// TestGetUsersHandler tests retrieving a list of users
func TestGetUsersHandler(t *testing.T) {
	suite := SetupSuite()
	suite.router.Handle("/users", MethodHandler("GET", WhichGetUsers(suite.handler.GetUsersHandler)))

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

		require.Equal(t, http.StatusOK, rr.Code)

		var responseBody []domain.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err)

		assert.Len(t, responseBody, len(sampleUsers))

		// Compare only relevant fields and allow slight timestamp variations
		for i, user := range responseBody {
			assert.Equal(t, sampleUsers[i].ID, user.ID, "handler returned incorrect ID for user %d", i)
			assert.Equal(t, sampleUsers[i].Name, user.Name, "handler returned incorrect Name for user %d", i)
			assert.Equal(t, sampleUsers[i].Email, user.Email, "handler returned incorrect Email for user %d", i)

			// Allow a margin for timestamp differences
			assert.WithinDuration(t, *sampleUsers[i].CreatedAt, *user.CreatedAt, time.Second, "handler returned incorrect CreatedAt for user %d", i)
			assert.WithinDuration(t, *sampleUsers[i].UpdatedAt, *user.UpdatedAt, time.Second, "handler returned incorrect UpdatedAt for user %d", i)
		}
	})

	t.Run("failure - no users found", func(t *testing.T) {
		// Mock service returning ErrNotFound
		suite.mockService.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return nil, common.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock service returning ErrInternalServerError
		suite.mockService.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return nil, common.ErrInternalServerError
		}

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rr.Body.String())
	})
}

// TestGetUserByEmailHandler tests retrieving a user by email
func TestGetUserByEmailHandler(t *testing.T) {
	suite := SetupSuite()
	suite.router.Handle("/users", MethodHandler("GET", WhichGetUsers(suite.handler.GetUserByEmailHandler)))

	// Generate a sample user
	sampleUser := testutils.GenerateMockUsers(1)[0]

	t.Run("success - user found", func(t *testing.T) {
		// Mock service returning the user
		suite.mockService.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return sampleUser, nil
		}

		// Create an HTTP request with the sample user's email
		req := httptest.NewRequest(http.MethodGet, "/users?email="+sampleUser.Email, nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var responseBody domain.User
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err)

		// Compare only essential fields
		assert.Equal(t, sampleUser.ID, responseBody.ID, "handler returned incorrect ID")
		assert.Equal(t, sampleUser.Name, responseBody.Name, "handler returned incorrect Name")
		assert.Equal(t, sampleUser.Email, responseBody.Email, "handler returned incorrect Email")

		// Allow minor differences in timestamps
		assert.WithinDuration(t, *sampleUser.CreatedAt, *responseBody.CreatedAt, time.Second, "handler returned incorrect CreatedAt")
		assert.WithinDuration(t, *sampleUser.UpdatedAt, *responseBody.UpdatedAt, time.Second, "handler returned incorrect UpdatedAt")
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

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", rr.Body.String())
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

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rr.Body.String())
	})
}

// TestUpdateUserHandler tests updating a user's information
func TestUpdateUserHandler(t *testing.T) {
	suite := SetupSuite()
	suite.router.Handle("/users/{id}", MethodHandler("PUT", VerifyUserID(suite.handler.UpdateUserHandler)))

	// Generate a sample user
	sampleUser := testutils.GenerateMockUsers(1)[0]

	t.Run("success - user updated", func(t *testing.T) {
		// Mock service returning updated user
		suite.mockService.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return domain.User{
				ID:        params.ID,
				Name:      params.Name,
				Email:     params.Email,
				CreatedAt: sampleUser.CreatedAt,
				UpdatedAt: common.Ptr(time.Now()),
			}, nil
		}

		// New data to update
		newName := "Updated John Doe"
		newEmail := "updated.johndoe@example.com"

		reqBody, err := json.Marshal(map[string]string{
			"name":  newName,
			"email": newEmail,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/users/"+sampleUser.ID.String(), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var responseBody domain.User
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err)

		// Compare only relevant fields
		assert.Equal(t, sampleUser.ID, responseBody.ID, "handler returned incorrect ID")
		assert.Equal(t, newName, responseBody.Name, "handler returned incorrect Name")
		assert.Equal(t, newEmail, responseBody.Email, "handler returned incorrect Email")

		// Allow minor differences in timestamps
		assert.WithinDuration(t, *sampleUser.CreatedAt, *responseBody.CreatedAt, time.Second, "handler returned incorrect CreatedAt")
		assert.WithinDuration(t, time.Now(), *responseBody.UpdatedAt, time.Second, "handler returned incorrect UpdatedAt")
	})

	t.Run("failure - user not found", func(t *testing.T) {
		suite.mockService.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		reqBody, _ := json.Marshal(map[string]string{"name": "Test Name", "email": "test@example.com"})
		req := httptest.NewRequest(http.MethodPut, "/users/"+uuid.New().String(), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", rr.Body.String())
	})

	t.Run("failure - invalid user ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/users/invalid-uuid", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusBadRequest)+"\n", rr.Body.String())
	})

	t.Run("failure - invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/users/"+sampleUser.ID.String(), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusBadRequest)+"\n", rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		suite.mockService.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return domain.User{}, common.ErrInternalServerError
		}

		reqBody, _ := json.Marshal(map[string]string{"name": "Updated Name", "email": "updated@example.com"})
		req := httptest.NewRequest(http.MethodPut, "/users/"+sampleUser.ID.String(), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rr.Body.String())
	})
}

// TestDeleteUserHandler tests deleting a user by ID
func TestDeleteUserHandler(t *testing.T) {
	suite := SetupSuite()
	suite.router.Handle("/users/{id}", MethodHandler("DELETE", VerifyUserID(suite.handler.DeleteUserHandler)))

	// Generate a sample user
	sampleUser := testutils.GenerateMockUsers(1)[0]

	t.Run("success - user deleted", func(t *testing.T) {
		// Mock service returning successful deletion
		suite.mockService.DeleteUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/users/"+sampleUser.ID.String(), nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("failure - user not found", func(t *testing.T) {
		// Mock service returning ErrNotFound
		suite.mockService.DeleteUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return common.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodDelete, "/users/"+uuid.New().String(), nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", rr.Body.String())
	})

	t.Run("failure - invalid user ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/invalid-uuid", nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusBadRequest)+"\n", rr.Body.String())
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock service returning an unexpected error
		suite.mockService.DeleteUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return fmt.Errorf("database error")
		}

		req := httptest.NewRequest(http.MethodDelete, "/users/"+sampleUser.ID.String(), nil)
		rr := httptest.NewRecorder()
		suite.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rr.Body.String())
	})
}
