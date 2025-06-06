package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/henryhall897/golang-todo-app/gen/mocks/usersmock"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/cache"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"github.com/henryhall897/golang-todo-app/internal/users/repository"
	"github.com/henryhall897/golang-todo-app/internal/users/testutils"
	rediswrapper "github.com/henryhall897/golang-todo-app/pkg/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Global test dependencies
type ServiceTestSuite struct {
	mockRepo *usersmock.RepositoryMock
	Redis    *RedisTestHelper
	Service  domain.Service
	ctx      context.Context
}

// SetupSuite initializes common dependencies
func SetupSuite() *ServiceTestSuite {
	logger := zap.NewNop() // No-op logger for tests

	mockRepo := &usersmock.RepositoryMock{}

	redis := SetupRedisTest()

	// Initialize Redis cache
	genericCache := rediswrapper.NewJSONCache(redis.Client, RedisTestPrefix, logger.Sugar())
	userCache := cache.NewRedisUser(genericCache)
	Service := New(mockRepo, userCache, logger.Sugar())

	return &ServiceTestSuite{
		mockRepo: mockRepo,
		Redis:    redis,
		Service:  Service,
		ctx:      context.Background(),
	}
}
func TestCreateUser(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	defer suite.Redis.Server.Close()

	// Define common test data
	ctx := context.Background()
	testUsers := testutils.GenerateMockUsers(1) // Use mock users generator
	testUserParams := domain.CreateUserParams{
		Name:  testUsers[0].Name,
		Email: testUsers[0].Email,
	}
	testUser := testUsers[0]

	t.Run("success - user created", func(t *testing.T) {
		// Mock successful user creation
		suite.mockRepo.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return testUser, nil
		}

		// Call the service method
		user, err := suite.Service.CreateUser(ctx, testUserParams)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("failure - email already exists", func(t *testing.T) {
		// Mock repository returning ErrEmailAlreadyExists
		suite.mockRepo.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return domain.User{}, repository.ErrEmailAlreadyExists
		}

		// Call the service method
		user, err := suite.Service.CreateUser(ctx, testUserParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrEmailAlreadyExists)) // Ensure correct sentinel
		assert.Equal(t, domain.User{}, user)                  // Should return an empty user
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock repository returning an unexpected error
		suite.mockRepo.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return domain.User{}, common.ErrInternalServerError
		}

		// Call the service method
		user, err := suite.Service.CreateUser(ctx, testUserParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Ensure correct sentinel
		assert.Equal(t, domain.User{}, user)                          // Should return an empty user
	})
}

func TestGetUserByID(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	defer suite.Redis.Server.Close()

	// Define common test data
	testUsers := testutils.GenerateMockUsers(1) // Use mock users generator
	testUser := testUsers[0]
	testUserID := testUser.ID

	t.Run("success - user found", func(t *testing.T) {
		// Mock successful user retrieval
		suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return testUser, nil
		}

		// Call the service method
		user, err := suite.Service.GetUserByID(suite.ctx, testUserID)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("failure - user not found", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrNotFound
		suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		// Call the service method
		user, err := suite.Service.GetUserByID(suite.ctx, testUserID)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrNotFound))
		assert.Equal(t, domain.User{}, user) // Should return an empty user
	})

	t.Run("failure - unexpected error", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning an unknown error
		suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, errors.New("database timeout")
		}

		// Call the service method
		user, err := suite.Service.GetUserByID(suite.ctx, testUserID)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Equal(t, domain.User{}, user)
	})
}

func TestGetUsers(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	defer suite.Redis.Server.Close()

	// Define common test data
	testUsers := testutils.GenerateMockUsers(3) // Use mock users generator
	testParams := domain.GetUsersParams{Limit: 10, Offset: 0}

	t.Run("success - users retrieved", func(t *testing.T) {
		// Mock successful user retrieval
		suite.mockRepo.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return testUsers, nil
		}

		// Call the service method
		users, err := suite.Service.GetUsers(suite.ctx, testParams)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, testUsers, users)
	})

	t.Run("failure - no users found", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrNotFound
		suite.mockRepo.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return []domain.User{}, common.ErrNotFound
		}

		// Call the service method
		users, err := suite.Service.GetUsers(suite.ctx, testParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrNotFound))
		assert.Empty(t, users) // Should return an empty list
	})

	t.Run("failure - invalid user data in DB", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrInvalidDbUserID
		suite.mockRepo.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return []domain.User{}, repository.ErrInvalidDbUserID
		}

		// Call the service method
		users, err := suite.Service.GetUsers(suite.ctx, testParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Empty(t, users)
	})

	t.Run("failure - failed to parse UUID", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrFailedToParseUUID
		suite.mockRepo.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return []domain.User{}, repository.ErrFailedToParseUUID
		}

		// Call the service method
		users, err := suite.Service.GetUsers(suite.ctx, testParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Empty(t, users)
	})

	t.Run("failure - unexpected error", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning an unknown error
		suite.mockRepo.GetUsersFunc = func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
			return []domain.User{}, errors.New("database timeout")
		}

		// Call the service method
		users, err := suite.Service.GetUsers(suite.ctx, testParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Empty(t, users)
	})
}

func TestGetUserByEmail(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	defer suite.Redis.Server.Close()

	// Define common test data
	testUsers := testutils.GenerateMockUsers(1) // Use mock users generator
	testUser := testUsers[0]
	testEmail := testUser.Email

	t.Run("success - user found", func(t *testing.T) {
		// Mock successful user retrieval
		suite.mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return testUser, nil
		}

		// Call the service method
		user, err := suite.Service.GetUserByEmail(suite.ctx, testEmail)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("failure - user not found", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrNotFound
		suite.mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		// Call the service method
		user, err := suite.Service.GetUserByEmail(suite.ctx, testEmail)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrNotFound))
		assert.Equal(t, domain.User{}, user) // Should return an empty user
	})

	t.Run("failure - invalid user data in DB", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrInvalidDbUserID
		suite.mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, repository.ErrInvalidDbUserID
		}

		// Call the service method
		user, err := suite.Service.GetUserByEmail(suite.ctx, testEmail)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Equal(t, domain.User{}, user)
	})

	t.Run("failure - failed to parse UUID", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrFailedToParseUUID
		suite.mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, repository.ErrFailedToParseUUID
		}

		// Call the service method
		user, err := suite.Service.GetUserByEmail(suite.ctx, testEmail)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Equal(t, domain.User{}, user)
	})

	t.Run("failure - unexpected error", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning an unknown error
		suite.mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, errors.New("database timeout")
		}

		// Call the service method
		user, err := suite.Service.GetUserByEmail(suite.ctx, testEmail)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Equal(t, domain.User{}, user)
	})
}

/*func TestGetUserByAuthID(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	defer suite.Redis.Server.Close()

	// Define common test data
	testUsers := testutils.GenerateMockUsers(1)
	testUser := testUsers[0]
	testAuthID := testUser.AuthID

	t.Run("success - user found", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis to test repo fallback

		// Mock repository to return valid user
		suite.mockRepo.GetUserByAuthIDFunc = func(ctx context.Context, authID string) (domain.User, error) {
			return testUser, nil
		}

		// Call the service method
		user, err := suite.Service.GetUserByAuthID(suite.ctx, testAuthID)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("failure - user not found", func(t *testing.T) {
		suite.Redis.Server.FlushAll()

		suite.mockRepo.GetUserByAuthIDFunc = func(ctx context.Context, authID string) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		user, err := suite.Service.GetUserByAuthID(suite.ctx, testAuthID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrNotFound))
		assert.Equal(t, domain.User{}, user)
	})

	t.Run("failure - invalid user ID", func(t *testing.T) {
		suite.Redis.Server.FlushAll()

		suite.mockRepo.GetUserByAuthIDFunc = func(ctx context.Context, authID string) (domain.User, error) {
			return domain.User{}, repository.ErrInvalidDbUserID
		}

		user, err := suite.Service.GetUserByAuthID(suite.ctx, testAuthID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError))
		assert.Equal(t, domain.User{}, user)
	})

	t.Run("failure - failed to parse UUID", func(t *testing.T) {
		suite.Redis.Server.FlushAll()

		suite.mockRepo.GetUserByAuthIDFunc = func(ctx context.Context, authID string) (domain.User, error) {
			return domain.User{}, repository.ErrFailedToParseUUID
		}

		user, err := suite.Service.GetUserByAuthID(suite.ctx, testAuthID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError))
		assert.Equal(t, domain.User{}, user)
	})

	t.Run("failure - unexpected error", func(t *testing.T) {
		suite.Redis.Server.FlushAll()

		suite.mockRepo.GetUserByAuthIDFunc = func(ctx context.Context, authID string) (domain.User, error) {
			return domain.User{}, errors.New("database timeout")
		}

		user, err := suite.Service.GetUserByAuthID(suite.ctx, testAuthID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError))
		assert.Equal(t, domain.User{}, user)
	})
}*/

func TestUpdateUser(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	defer suite.Redis.Server.Close()

	// Define common test data
	testUsers := testutils.GenerateMockUsers(1) // Use mock users generator
	testUser := testUsers[0]
	testUpdateParams := domain.UpdateUserParams{
		ID:    testUser.ID,
		Name:  "Updated Name",
		Email: "updated@example.com",
	}

	// Service still may do a getUserByID call to get the old user
	suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
		return testUser, nil
	}

	t.Run("success - user updated", func(t *testing.T) {
		// Mock successful user update
		suite.mockRepo.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return testUser, nil
		}

		// Call the service method
		updatedUser, err := suite.Service.UpdateUser(suite.ctx, testUpdateParams)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, testUser, updatedUser)
	})

	t.Run("failure - user not found", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrNotFound
		suite.mockRepo.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return domain.User{}, common.ErrNotFound
		}

		// Call the service method
		updatedUser, err := suite.Service.UpdateUser(suite.ctx, testUpdateParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrNotFound))
		assert.Equal(t, domain.User{}, updatedUser) // Should return an empty user
	})

	t.Run("failure - invalid user data in DB", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrInvalidDbUserID
		suite.mockRepo.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return domain.User{}, repository.ErrInvalidDbUserID
		}

		// Call the service method
		updatedUser, err := suite.Service.UpdateUser(suite.ctx, testUpdateParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Equal(t, domain.User{}, updatedUser)
	})

	t.Run("failure - failed to parse UUID", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning ErrFailedToParseUUID
		suite.mockRepo.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return domain.User{}, repository.ErrFailedToParseUUID
		}

		// Call the service method
		updatedUser, err := suite.Service.UpdateUser(suite.ctx, testUpdateParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Equal(t, domain.User{}, updatedUser)
	})

	t.Run("failure - unexpected error", func(t *testing.T) {
		suite.Redis.Server.FlushAll() // Clear Redis cache
		// Mock repository returning an unknown error
		suite.mockRepo.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return domain.User{}, errors.New("database timeout")
		}

		// Call the service method
		updatedUser, err := suite.Service.UpdateUser(suite.ctx, testUpdateParams)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
		assert.Equal(t, domain.User{}, updatedUser)
	})
}

func TestDeleteUser(t *testing.T) {
	suite := SetupSuite() // Load shared setup
	defer suite.Redis.Server.Close()

	// Define common test data
	testUser := testutils.GenerateMockUsers(1)[0] // Use mock user generator
	testUserID := testUser.ID

	// Mock repository for GetUserByID
	suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
		return testUser, nil
	}

	t.Run("success - user deleted", func(t *testing.T) {
		// Mock successful user deletion
		suite.mockRepo.DeleteUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		// Call the service method
		err := suite.Service.DeleteUser(suite.ctx, testUserID)

		// Assertions
		require.NoError(t, err)
	})

	t.Run("failure - user not found", func(t *testing.T) {
		// Mock repository returning ErrNotFound
		suite.mockRepo.DeleteUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return common.ErrNotFound
		}

		// Call the service method
		err := suite.Service.DeleteUser(suite.ctx, testUserID)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrNotFound))
	})

	t.Run("failure - internal server error", func(t *testing.T) {
		// Mock repository returning an unknown error
		suite.mockRepo.DeleteUserFunc = func(ctx context.Context, id uuid.UUID) error {
			return errors.New("database timeout")
		}

		// Call the service method
		err := suite.Service.DeleteUser(suite.ctx, testUserID)

		// Assertions
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrInternalServerError)) // Should be masked as internal error
	})
}
