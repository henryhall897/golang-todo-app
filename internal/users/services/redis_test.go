package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"github.com/henryhall897/golang-todo-app/internal/users/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RedisTestPrefix      = "test:"
	RedisTestEmailPrefix = "email:"
)

type RedisTestHelper struct {
	Server *miniredis.Miniredis
	Client *redis.Client
}

func RedisFullKey(subKey string) string {
	return RedisTestPrefix + ":" + subKey
}

func SetupRedisTest() *RedisTestHelper {
	// Start a new miniredis server
	server, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("Failed to start miniredis: %v", err))
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
		DB:   0,
	})

	return &RedisTestHelper{
		Server: server,
		Client: redisClient,
	}
}

func TestRedisConnection(t *testing.T) {
	redisHelper := SetupRedisTest()
	defer redisHelper.Server.Close() // Cleanup Miniredis after test

	// Test setting a value
	err := redisHelper.Client.Set(context.Background(), "testKey", "testValue", 0).Err()
	require.NoError(t, err, "Failed to set value in Miniredis")

	// Test getting the value
	val, err := redisHelper.Client.Get(context.Background(), "testKey").Result()
	require.NoError(t, err, "Failed to get value from Miniredis")
	assert.Equal(t, "testValue", val, "Redis value mismatch")
}

func TestUserCache_Initialization(t *testing.T) {
	suite := SetupSuite()
	require.NotNil(t, suite.Service, "Service should not be nil")

	// Cast Service to the concrete type (if possible)
	serviceImpl, ok := suite.Service.(*service)
	require.True(t, ok, "Service should be of concrete type *service")
	require.NotNil(t, serviceImpl.cache, "UserCache inside service should not be nil")
}

func TestCreateUser_Cache(t *testing.T) {
	suite := SetupSuite()            // Load shared test setup
	defer suite.Redis.Server.Close() // Cleanup Miniredis after test

	mockUsers := testutils.GenerateMockUsers(1) // Generate test user
	testUser := mockUsers[0]

	cacheKeyByID := domain.CacheKeyByID(testUser.ID)
	cacheKeyByID = RedisFullKey(cacheKeyByID)
	cacheKeyByEmail := domain.CacheKeyByEmail(testUser.Email)
	cacheKeyByEmail = RedisFullKey(cacheKeyByEmail)

	t.Run("success - user created and cached", func(t *testing.T) {
		// Ensure cache does not exist before creation
		require.False(t, suite.Redis.Server.Exists(cacheKeyByID), "User ID cache should not exist before creation")
		require.False(t, suite.Redis.Server.Exists(cacheKeyByEmail), "User email cache should not exist before creation")

		// Mock DB creation
		suite.mockRepo.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return testUser, nil
		}

		// Call service method (should create user and store in Redis)
		_, err := suite.Service.CreateUser(suite.ctx, domain.CreateUserParams{
			Name:   testUser.Name,
			Email:  testUser.Email,
			AuthID: testUser.AuthID,
		})
		require.NoError(t, err)
		fmt.Printf("Current Redis keys:\n")
		for _, key := range suite.Redis.Server.Keys() {
			val, _ := suite.Redis.Server.Get(key)
			fmt.Printf("  %s = %s\n", key, val)
		}
		fmt.Println(cacheKeyByID)
		fmt.Println(cacheKeyByEmail)

		// Verify user is now cached in Redis by ID
		cachedUserJSONByID, err := suite.Redis.Server.Get(cacheKeyByID)
		require.NoError(t, err, "Expected Redis to contain the cached user by ID")
		assert.NotEmpty(t, cachedUserJSONByID, "Expected Redis to store the user after creation")

		// Deserialize only the full user record (from ID key)
		var cachedUserByID domain.User
		require.NoError(t, json.Unmarshal([]byte(cachedUserJSONByID), &cachedUserByID))

		// Ensure cached users match the created user
		assert.Equal(t, testUser.ID, cachedUserByID.ID)
		assert.Equal(t, testUser.Name, cachedUserByID.Name)
		assert.Equal(t, testUser.Email, cachedUserByID.Email)

		// Check email pointer exists and points to correct ID
		emailPointerValue, err := suite.Redis.Server.Get(cacheKeyByEmail)
		require.NoError(t, err, "Expected Redis to contain the email pointer")
		parsedEmailUUID, err := uuid.Parse(emailPointerValue)
		require.NoError(t, err, "Email pointer is not a valid UUID")
		assert.Equal(t, testUser.ID, parsedEmailUUID, "Email pointer should match user ID")

		// Check auth_id pointer exists and points to correct ID
		authIDPointerKey := domain.CacheKeyByAuthID(testUser.AuthID)
		cacheKeyByAuthID := RedisFullKey(authIDPointerKey)

		authPointerValue, err := suite.Redis.Server.Get(cacheKeyByAuthID)
		require.NoError(t, err, "Expected Redis to contain the auth_id pointer")
		parsedAuthUUID, err := uuid.Parse(authPointerValue)
		require.NoError(t, err, "AuthID pointer is not a valid UUID")
		assert.Equal(t, testUser.ID, parsedAuthUUID, "AuthID pointer should match user ID")

	})

	t.Run("failure - Redis error does not impact user creation", func(t *testing.T) {
		// Simulate Redis failure by stopping MiniRedis
		suite.Redis.Server.Close()

		// Mock DB creation
		suite.mockRepo.CreateUserFunc = func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
			return testUser, nil
		}

		// Call service method (should succeed even if Redis fails)
		_, err := suite.Service.CreateUser(suite.ctx, domain.CreateUserParams{
			Name:  testUser.Name,
			Email: testUser.Email,
		})
		require.NoError(t, err, "User creation should succeed even if Redis fails")
	})
}

func TestGetUserByID_Cache(t *testing.T) {
	suite := SetupSuite()            // Load shared test setup
	defer suite.Redis.Server.Close() // Cleanup Miniredis after test

	mockUsers := testutils.GenerateMockUsers(1)
	testUser := mockUsers[0]
	cacheKey := domain.CacheKeyByID(testUser.ID)
	cacheKey = RedisFullKey(cacheKey)

	t.Run("success - cache miss, fetch from DB and store in Redis", func(t *testing.T) {
		// Ensure key does not exist before DB call (cache miss)
		require.False(t, suite.Redis.Server.Exists(cacheKey), "Key should not exist in Redis before DB fetch")

		// Mock DB fetch
		suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return testUser, nil
		}

		// Call service method (should fetch from DB and store in Redis)
		_, err := suite.Service.GetUserByID(suite.ctx, testUser.ID)
		require.NoError(t, err)

		// Verify the user is now cached in Redis
		cachedUserJSON, err := suite.Redis.Server.Get(cacheKey)
		require.NoError(t, err, "Expected Redis to contain the cached user")
		assert.NotEmpty(t, cachedUserJSON, "Expected Redis to store the user after DB fetch")

		// Deserialize cached user
		var cachedUser domain.User
		err = json.Unmarshal([]byte(cachedUserJSON), &cachedUser)
		require.NoError(t, err, "Failed to deserialize cached user from Redis")

		// Ensure cached user matches the original user
		assert.Equal(t, testUser.ID, cachedUser.ID)
		assert.Equal(t, testUser.Name, cachedUser.Name)
		assert.Equal(t, testUser.Email, cachedUser.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *cachedUser.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *cachedUser.UpdatedAt, time.Millisecond)
	})

	t.Run("success - cache hit", func(t *testing.T) {
		// Serialize and manually store user in Redis (simulating a previous cache)
		userJSON, _ := json.Marshal(testUser)
		suite.Redis.Server.Set(cacheKey, string(userJSON))

		// Call service method (should retrieve from Redis instead of DB)
		user, err := suite.Service.GetUserByID(suite.ctx, testUser.ID)
		require.NoError(t, err)

		// Ensure DB was not called
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Name, user.Name)
		assert.Equal(t, testUser.Email, user.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *user.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *user.UpdatedAt, time.Millisecond)
	})

	t.Run("failure - Redis error, fallback to DB", func(t *testing.T) {
		// Simulate Redis failure by stopping MiniRedis
		suite.Redis.Server.Close()

		// Expect DB fetch since Redis is down
		suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return testUser, nil
		}

		// Call service method (should fetch from DB due to Redis failure)
		user, err := suite.Service.GetUserByID(suite.ctx, testUser.ID)

		// Ensure service still functions correctly without Redis
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Name, user.Name)
		assert.Equal(t, testUser.Email, user.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *user.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *user.UpdatedAt, time.Millisecond)
	})
}

func TestGetUsers_Cache(t *testing.T) {
	suite := SetupSuite()            // Load shared test setup
	defer suite.Redis.Server.Close() // Cleanup Miniredis after test

	mockUsers := testutils.GenerateMockUsers(3) // Generate test users
	params := domain.GetUsersParams{Limit: 3, Offset: 0}
	key := domain.CacheKeyByPagination(params.Limit, params.Offset)
	cacheKey := RedisFullKey(key)

	t.Run("success - cache miss, fetch from DB and store in Redis", func(t *testing.T) {
		// Ensure key does not exist before DB call (cache miss)
		require.False(t, suite.Redis.Server.Exists(cacheKey), "Key should not exist in Redis before DB fetch")

		// Mock DB fetch
		suite.mockRepo.GetUsersFunc = func(ctx context.Context, p domain.GetUsersParams) ([]domain.User, error) {
			return mockUsers, nil
		}

		// Call service method (should fetch from DB and store in Redis)
		_, err := suite.Service.GetUsers(suite.ctx, params)
		require.NoError(t, err)

		// Verify users are now cached in Redis
		cachedUsersJSON, err := suite.Redis.Server.Get(cacheKey)
		require.NoError(t, err, "Expected Redis to contain cached users")
		assert.NotEmpty(t, cachedUsersJSON, "Expected Redis to store users after DB fetch")

		// Deserialize cached users
		var cachedUsers []domain.User
		err = json.Unmarshal([]byte(cachedUsersJSON), &cachedUsers)
		require.NoError(t, err, "Failed to deserialize cached users from Redis")

		// Ensure cached users match the original users
		assert.Equal(t, len(mockUsers), len(cachedUsers))
		for i, user := range cachedUsers {
			assert.Equal(t, mockUsers[i].ID, user.ID)
			assert.Equal(t, mockUsers[i].Name, user.Name)
			assert.Equal(t, mockUsers[i].Email, user.Email)
			assert.WithinDuration(t, *mockUsers[i].CreatedAt, *user.CreatedAt, time.Millisecond)
			assert.WithinDuration(t, *mockUsers[i].UpdatedAt, *user.UpdatedAt, time.Millisecond)
		}
	})

	t.Run("success - cache hit", func(t *testing.T) {
		// Serialize and manually store users in Redis (simulating a previous cache)
		usersJSON, _ := json.Marshal(mockUsers)
		suite.Redis.Server.Set(cacheKey, string(usersJSON))

		// Call service method (should retrieve from Redis instead of DB)
		users, err := suite.Service.GetUsers(suite.ctx, params)
		require.NoError(t, err)

		// Ensure retrieved users match cached users
		assert.Equal(t, len(mockUsers), len(users))
		for i, user := range users {
			assert.Equal(t, mockUsers[i].ID, user.ID)
			assert.Equal(t, mockUsers[i].Name, user.Name)
			assert.Equal(t, mockUsers[i].Email, user.Email)
			assert.WithinDuration(t, *mockUsers[i].CreatedAt, *user.CreatedAt, time.Millisecond)
			assert.WithinDuration(t, *mockUsers[i].UpdatedAt, *user.UpdatedAt, time.Millisecond)
		}
	})

	t.Run("failure - Redis error, fallback to DB", func(t *testing.T) {
		// Simulate Redis failure by stopping MiniRedis
		suite.Redis.Server.Close()

		// Expect DB fetch since Redis is down
		suite.mockRepo.GetUsersFunc = func(ctx context.Context, p domain.GetUsersParams) ([]domain.User, error) {
			return mockUsers, nil
		}

		// Call service method (should fetch from DB due to Redis failure)
		users, err := suite.Service.GetUsers(suite.ctx, params)

		// Ensure service still functions correctly without Redis
		require.NoError(t, err)
		assert.Equal(t, len(mockUsers), len(users))
		for i, user := range users {
			assert.Equal(t, mockUsers[i].ID, user.ID)
			assert.Equal(t, mockUsers[i].Name, user.Name)
			assert.Equal(t, mockUsers[i].Email, user.Email)
			assert.WithinDuration(t, *mockUsers[i].CreatedAt, *user.CreatedAt, time.Millisecond)
			assert.WithinDuration(t, *mockUsers[i].UpdatedAt, *user.UpdatedAt, time.Millisecond)
		}
	})
}

func TestGetUserByEmail_Cache(t *testing.T) {
	suite := SetupSuite()            // Load shared test setup
	defer suite.Redis.Server.Close() // Cleanup Miniredis after test

	mockUsers := testutils.GenerateMockUsers(1) // Generate test user
	testUser := mockUsers[0]
	key := domain.CacheKeyByEmail(testUser.Email)
	cacheKey := RedisFullKey(key)

	t.Run("success - cache miss, fetch from DB and store in Redis", func(t *testing.T) {
		// Ensure key does not exist before DB call (cache miss)
		require.False(t, suite.Redis.Server.Exists(cacheKey), "Key should not exist in Redis before DB fetch")

		// Mock DB fetch
		suite.mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return testUser, nil
		}

		// Call service method (should fetch from DB and store in Redis)
		_, err := suite.Service.GetUserByEmail(suite.ctx, testUser.Email)
		fullUserKey := RedisFullKey(domain.CacheKeyByID(testUser.ID))
		require.NoError(t, err)

		// Verify the user is now cached in Redis
		cachedUserJSON, err := suite.Redis.Server.Get(fullUserKey)
		require.NoError(t, err, "Expected Redis to contain the cached user")
		assert.NotEmpty(t, cachedUserJSON, "Expected Redis to store the user after DB fetch")

		// Deserialize cached user
		var cachedUser domain.User
		err = json.Unmarshal([]byte(cachedUserJSON), &cachedUser)
		require.NoError(t, err, "Failed to deserialize cached user from Redis")

		// Ensure cached user matches the original user
		assert.Equal(t, testUser.ID, cachedUser.ID)
		assert.Equal(t, testUser.Name, cachedUser.Name)
		assert.Equal(t, testUser.Email, cachedUser.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *cachedUser.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *cachedUser.UpdatedAt, time.Millisecond)
	})

	t.Run("success - cache hit", func(t *testing.T) {
		// Serialize and manually store user in Redis (simulating a previous cache)
		userJSON, _ := json.Marshal(testUser)
		suite.Redis.Server.Set(cacheKey, string(userJSON))

		// Call service method (should retrieve from Redis instead of DB)
		user, err := suite.Service.GetUserByEmail(suite.ctx, testUser.Email)
		require.NoError(t, err)

		// Ensure retrieved user matches cached user
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Name, user.Name)
		assert.Equal(t, testUser.Email, user.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *user.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *user.UpdatedAt, time.Millisecond)
	})

	t.Run("failure - Redis error, fallback to DB", func(t *testing.T) {
		// Simulate Redis failure by stopping MiniRedis
		suite.Redis.Server.Close()

		// Expect DB fetch since Redis is down
		suite.mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (domain.User, error) {
			return testUser, nil
		}

		// Call service method (should fetch from DB due to Redis failure)
		user, err := suite.Service.GetUserByEmail(suite.ctx, testUser.Email)

		// Ensure service still functions correctly without Redis
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Name, user.Name)
		assert.Equal(t, testUser.Email, user.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *user.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *user.UpdatedAt, time.Millisecond)
	})
}

func TestGetUserByAuthID_Cache(t *testing.T) {
	suite := SetupSuite()
	defer suite.Redis.Server.Close()

	mockUsers := testutils.GenerateMockUsers(1)
	testUser := mockUsers[0]

	// Cache key for the pointer (auth_id → id)
	pointerKey := domain.CacheKeyByAuthID(testUser.AuthID)
	fullKey := domain.CacheKeyByID(testUser.ID)
	pointerRedisKey := RedisFullKey(pointerKey)
	fullRedisKey := RedisFullKey(fullKey)

	t.Run("success - cache miss, fetch from DB and store pointer + user", func(t *testing.T) {
		require.False(t, suite.Redis.Server.Exists(pointerRedisKey), "Pointer key should not exist before call")

		// Simulate DB returning the user
		suite.mockRepo.GetUserByAuthIDFunc = func(ctx context.Context, authID string) (domain.User, error) {
			return testUser, nil
		}

		_, err := suite.Service.GetUserByAuthID(suite.ctx, testUser.AuthID)
		require.NoError(t, err)

		// Verify pointer is now in Redis
		pointerVal, err := suite.Redis.Server.Get(pointerRedisKey)
		require.NoError(t, err, "Expected Redis to contain pointer key")
		assert.Equal(t, fullKey, pointerVal)

		// Verify full user is cached
		userJSON, err := suite.Redis.Server.Get(fullRedisKey)
		require.NoError(t, err, "Expected Redis to contain full user")
		assert.NotEmpty(t, userJSON)

		// Validate deserialization
		var cachedUser domain.User
		err = json.Unmarshal([]byte(userJSON), &cachedUser)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, cachedUser.ID)
		assert.Equal(t, testUser.AuthID, cachedUser.AuthID)
		assert.Equal(t, testUser.Email, cachedUser.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *cachedUser.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *cachedUser.UpdatedAt, time.Millisecond)
	})

	t.Run("success - cache hit via pointer", func(t *testing.T) {
		userJSON, _ := json.Marshal(testUser)

		// Set the full user and pointer manually in Redis
		suite.Redis.Server.Set(pointerRedisKey, fullKey)
		suite.Redis.Server.Set(fullRedisKey, string(userJSON))

		user, err := suite.Service.GetUserByAuthID(suite.ctx, testUser.AuthID)
		require.NoError(t, err)

		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.AuthID, user.AuthID)
		assert.Equal(t, testUser.Email, user.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *user.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *user.UpdatedAt, time.Millisecond)
	})

	t.Run("failure - Redis down, fallback to DB", func(t *testing.T) {
		suite.Redis.Server.Close() // simulate Redis failure

		// Mock DB fallback
		suite.mockRepo.GetUserByAuthIDFunc = func(ctx context.Context, authID string) (domain.User, error) {
			return testUser, nil
		}

		user, err := suite.Service.GetUserByAuthID(suite.ctx, testUser.AuthID)
		require.NoError(t, err)

		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.AuthID, user.AuthID)
		assert.Equal(t, testUser.Email, user.Email)
		assert.WithinDuration(t, *testUser.CreatedAt, *user.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *testUser.UpdatedAt, *user.UpdatedAt, time.Millisecond)
	})
}

func TestUpdateUser_Cache(t *testing.T) {
	suite := SetupSuite()            // Load shared test setup
	defer suite.Redis.Server.Close() // Cleanup Miniredis after test

	mockUsers := testutils.GenerateMockUsers(1) // Generate a test user
	testUser := mockUsers[0]
	updateParams := domain.UpdateUserParams{
		ID:    testUser.ID,
		Name:  "Updated Name",
		Email: "updated@example.com",
	}

	// Cache keys based on OLD values before update
	cacheKeyByID := domain.CacheKeyByID(testUser.ID)
	cacheKeyByID = RedisFullKey(cacheKeyByID)
	cacheKeyByOldEmail := domain.CacheKeyByEmail(testUser.Email)
	cacheKeyByOldEmail = RedisFullKey(cacheKeyByOldEmail)

	// Service still may do a getUserByID call to get the old user
	suite.mockRepo.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (domain.User, error) {
		return testUser, nil
	}

	t.Run("success - update user, refresh cache with updated data", func(t *testing.T) {
		// Cache old user details
		oldUserJSON, _ := json.Marshal(testUser)
		suite.Redis.Server.Set(cacheKeyByID, string(oldUserJSON))
		suite.Redis.Server.Set(cacheKeyByOldEmail, string(oldUserJSON))

		// Ensure keys exist before update
		require.True(t, suite.Redis.Server.Exists(cacheKeyByID), "User should be cached by ID before update")
		require.True(t, suite.Redis.Server.Exists(cacheKeyByOldEmail), "User should be cached by email before update")

		// Mock DB update
		updatedUser := testUser
		updatedUser.Name = updateParams.Name
		updatedUser.Email = updateParams.Email
		suite.mockRepo.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return updatedUser, nil
		}

		fmt.Println("Before update: Old email cache exists?", suite.Redis.Server.Exists(cacheKeyByOldEmail))
		fmt.Println("Attempting to delete keys:", cacheKeyByID, cacheKeyByOldEmail)
		// Call service method
		_, err := suite.Service.UpdateUser(suite.ctx, updateParams)
		require.NoError(t, err)
		fmt.Println("After update: Old email cache exists?", suite.Redis.Server.Exists(cacheKeyByOldEmail))
		/*suite.Redis.Server.Del(cacheKeyByOldEmail)
		fmt.Println("After delete: Old email cache exists?", suite.Redis.Server.Exists(cacheKeyByOldEmail))*/
		// Verify old email key was removed
		require.False(t, suite.Redis.Server.Exists(cacheKeyByOldEmail), "Old email cache should be deleted after update")

		// Verify the updated user is now cached under the same ID key
		cachedUserJSON, err := suite.Redis.Server.Get(cacheKeyByID)
		require.NoError(t, err, "Expected Redis to contain the updated user")
		assert.NotEmpty(t, cachedUserJSON, "Expected Redis to store the updated user after DB update")

		// Deserialize cached user
		var cachedUser domain.User
		err = json.Unmarshal([]byte(cachedUserJSON), &cachedUser)
		require.NoError(t, err, "Failed to deserialize cached updated user from Redis")

		// Ensure cached user matches updated user
		assert.Equal(t, updatedUser.ID, cachedUser.ID)       // ID should remain the same
		assert.Equal(t, updatedUser.Name, cachedUser.Name)   // Name should be updated
		assert.Equal(t, updatedUser.Email, cachedUser.Email) // Email should be updated
		assert.WithinDuration(t, *updatedUser.CreatedAt, *cachedUser.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *updatedUser.UpdatedAt, *cachedUser.UpdatedAt, time.Millisecond)
	})

	t.Run("failure - Redis error, still updates DB", func(t *testing.T) {
		// Simulate Redis failure by stopping MiniRedis
		suite.Redis.Server.Close()

		// Expect DB update
		updatedUser := testUser
		updatedUser.Name = updateParams.Name
		updatedUser.Email = updateParams.Email
		suite.mockRepo.UpdateUserFunc = func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
			return updatedUser, nil
		}

		// Call service method
		user, err := suite.Service.UpdateUser(suite.ctx, updateParams)

		// Ensure service still functions correctly without Redis
		require.NoError(t, err)
		assert.Equal(t, updatedUser.ID, user.ID)
		assert.Equal(t, updatedUser.Name, user.Name)
		assert.Equal(t, updatedUser.Email, user.Email)
		assert.WithinDuration(t, *updatedUser.CreatedAt, *user.CreatedAt, time.Millisecond)
		assert.WithinDuration(t, *updatedUser.UpdatedAt, *user.UpdatedAt, time.Millisecond)
	})
}
