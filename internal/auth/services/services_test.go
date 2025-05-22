package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	//internal packages
	jwt "github.com/golang-jwt/jwt/v5"
	authmock "github.com/henryhall897/golang-todo-app/gen/mocks/authmocks"
	usersmock "github.com/henryhall897/golang-todo-app/gen/mocks/usersmock"
	"github.com/henryhall897/golang-todo-app/internal/auth/domain"
	"github.com/henryhall897/golang-todo-app/internal/auth/testutils"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	udomain "github.com/henryhall897/golang-todo-app/internal/users/domain"
	utestutils "github.com/henryhall897/golang-todo-app/internal/users/testutils"
	jdomain "github.com/henryhall897/golang-todo-app/pkg/jwt/domain"
	token "github.com/henryhall897/golang-todo-app/pkg/jwt/token"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepo    *authmock.RepositoryMock
	JWT         jdomain.TokenGenerator
	mockUserSvc *usersmock.ServiceMock
	Cache       *authmock.CacheMock
	Service     domain.Service
	ctx         context.Context
}

func TestAuth(t *testing.T) {
	suite.Run(t, &ServiceTestSuite{})
}

func (s *ServiceTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.JWT = token.NewJWTTokenGenerator(token.TokenConfig{
		SecretKey:     "test-secret",
		Issuer:        "test-suite",
		TokenDuration: time.Hour,
	})
}

func (s *ServiceTestSuite) SetupTest() {
	// Recreate mocks for isolation per test
	s.mockRepo = &authmock.RepositoryMock{}
	s.mockUserSvc = &usersmock.ServiceMock{}
	s.Cache = &authmock.CacheMock{}

	s.Service = New(
		s.mockRepo,
		s.Cache,
		s.mockUserSvc,
		s.JWT,
		zap.NewNop().Sugar(),
	)
}

func (s *ServiceTestSuite) TearDownTest() {
	// Optionally clear mock state or reset counters
	s.mockRepo = nil
	s.mockUserSvc = nil
	s.Cache = nil
	s.Service = nil
}

func (suite *ServiceTestSuite) TestLoginOrRegister_Scenarios() {
	t := suite.T()
	// Generate consistent test user and auth input
	mockUsers := utestutils.GenerateMockUsers(1)
	testUser := mockUsers[0]

	mockAuthParams := testutils.GenerateMockAuthParams(mockUsers)
	authInput := domain.AuthLoginParams{
		AuthID:   mockAuthParams[0].AuthID,
		Provider: mockAuthParams[0].Provider,
		Email:    testUser.Email,
		Name:     testUser.Name,
	}

	t.Run("Success - existing auth identity (login path)", func(t *testing.T) {
		suite.mockRepo.GetAuthIdentityByAuthIDFunc = func(ctx context.Context, authID string) (domain.AuthIdentity, error) {
			return domain.AuthIdentity{
				AuthID:   authInput.AuthID,
				Provider: authInput.Provider,
				UserID:   testUser.ID,
			}, nil
		}

		suite.mockUserSvc.GetUserByIDFunc = func(ctx context.Context, id uuid.UUID) (udomain.User, error) {
			return testUser, nil
		}

		token, user, err := suite.Service.LoginOrRegister(suite.ctx, authInput)

		require.NoError(t, err)
		assert.Equal(t, testUser, user)

		claims, err := suite.JWT.Parse(suite.ctx, token)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, claims.UserID)
		assert.Equal(t, testUser.Role, claims.Role)
		assert.Equal(t, "test-suite", claims.Issuer)
	})

	t.Run("Success - new user registration", func(t *testing.T) {
		suite.mockRepo.GetAuthIdentityByAuthIDFunc = func(ctx context.Context, authID string) (domain.AuthIdentity, error) {
			return domain.AuthIdentity{}, common.ErrNotFound
		}

		suite.mockUserSvc.CreateUserFunc = func(ctx context.Context, params udomain.CreateUserParams) (udomain.User, error) {
			return testUser, nil
		}

		suite.mockRepo.CreateAuthIdentityFunc = func(ctx context.Context, params domain.CreateAuthIdentityParams) (domain.AuthIdentity, error) {
			return domain.AuthIdentity{
				AuthID:   authInput.AuthID,
				Provider: authInput.Provider,
				UserID:   testUser.ID,
			}, nil
		}

		token, user, err := suite.Service.LoginOrRegister(suite.ctx, authInput)

		require.NoError(t, err)
		assert.Equal(t, testUser, user)

		claims, err := suite.JWT.Parse(suite.ctx, token)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, claims.UserID)
		assert.Equal(t, testUser.Role, claims.Role)
	})
}

func (suite *ServiceTestSuite) TestLogout_Scenarios() {
	t := suite.T()
	testUser := utestutils.GenerateMockUsers(1)[0]

	t.Run("Success - valid token is blacklisted", func(t *testing.T) {
		// Generate token
		token, err := suite.JWT.Gen(suite.ctx, jdomain.Payload{
			ID:   testUser.ID,
			Role: testUser.Role,
		})
		require.NoError(t, err)

		// Parse to extract jti + exp for TTL verification
		claims, err := suite.JWT.Parse(suite.ctx, token)
		require.NoError(t, err)

		expectedJTI := claims.ID
		expectedTTL := time.Until(claims.ExpiresAt.Time)

		var blacklistedKey string
		var blacklistedTTL time.Duration

		suite.Cache.BlacklistTokenFunc = func(ctx context.Context, jti string, ttl time.Duration) error {
			blacklistedKey = jti
			blacklistedTTL = ttl
			return nil
		}

		err = suite.Service.Logout(suite.ctx, token)
		require.NoError(t, err)

		assert.Equal(t, expectedJTI, blacklistedKey, "Expected jti to be blacklisted")
		assert.WithinDuration(t, time.Now().Add(expectedTTL), time.Now().Add(blacklistedTTL), 2*time.Second, "TTL should match token expiration")
	})

	t.Run("Skip - token already expired", func(t *testing.T) {
		// Issue a token with past expiration
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jdomain.Claims{
			UserID: testUser.ID,
			Role:   testUser.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "test-suite",
				Subject:   testUser.ID.String(),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				ID:        uuid.NewString(),
			},
		})
		signedExpiredToken, err := expiredToken.SignedString([]byte("test-secret"))
		require.NoError(t, err)

		// Ensure BlacklistToken is NOT called
		suite.Cache.BlacklistTokenFunc = func(ctx context.Context, jti string, ttl time.Duration) error {
			t.Errorf("BlacklistToken should not be called for expired token")
			return nil
		}

		err = suite.Service.Logout(suite.ctx, signedExpiredToken)
		require.NoError(t, err)
	})

	t.Run("Failure - invalid token format", func(t *testing.T) {
		err := suite.Service.Logout(suite.ctx, "this.is.not.a.valid.jwt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})
}
