package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/repository"
	"github.com/nurashi/car-rental-identity/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) UpdateEmailVerified(ctx context.Context, userID string, verified bool) error {
	args := m.Called(ctx, userID, verified)
	return args.Error(0)
}

func (m *mockUserRepo) List(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *mockUserRepo) BlockUser(ctx context.Context, userID, reason string) error {
	args := m.Called(ctx, userID, reason)
	return args.Error(0)
}

type mockEmailVerRepo struct {
	mock.Mock
}

func (m *mockEmailVerRepo) Create(ctx context.Context, ev *domain.EmailVerification) error {
	args := m.Called(ctx, ev)
	return args.Error(0)
}

func (m *mockEmailVerRepo) GetByToken(ctx context.Context, token string) (*domain.EmailVerification, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EmailVerification), args.Error(1)
}

func (m *mockEmailVerRepo) MarkUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEmailVerRepo) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockRedisDB struct {
	mock.Mock
}

func (m *mockRedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return m.Called(ctx, key, value, expiration).Error(0)
}

func (m *mockRedisDB) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockRedisDB) Delete(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	userRepo := new(mockUserRepo)
	emailVerRepo := new(mockEmailVerRepo)

	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(&domain.User{ID: "1", Email: "test@example.com"}, nil)

	authService := service.NewAuthService(userRepo, emailVerRepo, service.Config{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 24,
	})

	userID, err := authService.Register(context.Background(), "test@example.com", "password", "John", "Doe", "1234567890")

	assert.Empty(t, userID)
	assert.Equal(t, service.ErrUserAlreadyExists, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	emailVerRepo := new(mockEmailVerRepo)

	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), repository.ErrUserNotFound)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Run(func(args mock.Arguments) {
		user := args.Get(1).(*domain.User)
		user.ID = "test-user-123"
	}).Return(nil)
	emailVerRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.EmailVerification")).Return(nil)

	authService := service.NewAuthService(userRepo, emailVerRepo, service.Config{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 24,
	})

	userID, err := authService.Register(context.Background(), "test@example.com", "password123", "John", "Doe", "1234567890")

	assert.Equal(t, "test-user-123", userID)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	emailVerRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userRepo := new(mockUserRepo)
	emailVerRepo := new(mockEmailVerRepo)

	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return((*domain.User)(nil), repository.ErrUserNotFound)

	authService := service.NewAuthService(userRepo, emailVerRepo, service.Config{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 24,
	})

	user, token, err := authService.Login(context.Background(), "test@example.com", "wrongpassword")

	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, service.ErrInvalidCredentials, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_BlockedUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	emailVerRepo := new(mockEmailVerRepo)

	hashed, _ := service.HashPasswordForTest("password")
	blockedUser := &domain.User{
		ID:           "1",
		Email:        "test@example.com",
		PasswordHash: hashed,
		Status:       domain.UserStatusBlocked,
	}
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(blockedUser, nil)

	authService := service.NewAuthService(userRepo, emailVerRepo, service.Config{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 24,
	})

	user, token, err := authService.Login(context.Background(), "test@example.com", "password")

	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, service.ErrUserBlocked, err)
}

func TestAuthService_GenerateAndValidateToken(t *testing.T) {
	userRepo := new(mockUserRepo)
	emailVerRepo := new(mockEmailVerRepo)

	authService := service.NewAuthService(userRepo, emailVerRepo, service.Config{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 24,
	})

	token, err := authService.GenerateJWTForTest("user-123", "test@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}
