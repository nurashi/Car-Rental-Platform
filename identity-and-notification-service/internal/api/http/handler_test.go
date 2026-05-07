package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httphandler "github.com/nurashi/car-rental-identity/internal/api/http"
	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, email, password, firstName, lastName, phone string) (string, error) {
	args := m.Called(ctx, email, password, firstName, lastName, phone)
	return args.String(0), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *mockAuthService) VerifyEmail(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAuthService) ResendVerification(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAuthService) ValidateToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func TestRegisterHandler_Success(t *testing.T) {
	authService := new(mockAuthService)
	userService := new(mockUserService)

	authService.On("Register", mock.Anything, "test@example.com", "password123", "John", "Doe", "123456").Return("user-1", nil)

	handler := httphandler.NewHandler(authService, userService)

	body := map[string]string{
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "John",
		"last_name":  "Doe",
		"phone":      "123456",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "user-1", resp["user_id"])
	authService.AssertExpectations(t)
}

func TestLoginHandler_Success(t *testing.T) {
	authService := new(mockAuthService)
	userService := new(mockUserService)

	user := &domain.User{ID: "1", Email: "test@example.com", FirstName: "John"}
	authService.On("Login", mock.Anything, "test@example.com", "password123").Return(user, "token-123", nil)

	handler := httphandler.NewHandler(authService, userService)

	body := map[string]string{"email": "test@example.com", "password": "password123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "token-123", resp["token"])
	authService.AssertExpectations(t)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	authService := new(mockAuthService)
	userService := new(mockUserService)

	authService.On("Login", mock.Anything, "test@example.com", "wrong").Return((*domain.User)(nil), "", service.ErrInvalidCredentials)

	handler := httphandler.NewHandler(authService, userService)

	body := map[string]string{"email": "test@example.com", "password": "wrong"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProfileHandler_Success(t *testing.T) {
	authService := new(mockAuthService)
	userService := new(mockUserService)

	user := &domain.User{ID: "1", Email: "test@example.com", FirstName: "John"}
	userService.On("GetProfile", mock.Anything, "1").Return(user, nil)

	handler := httphandler.NewHandler(authService, userService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/1", nil)
	req.Header.Set("Authorization", "valid-token")
	authService.On("ValidateToken", "valid-token").Return("1", nil)
	w := httptest.NewRecorder()

	handler.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	userService.AssertExpectations(t)
}

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) UpdateProfile(ctx context.Context, input service.UpdateProfileInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *mockUserService) BlockUser(ctx context.Context, userID, reason string) error {
	args := m.Called(ctx, userID, reason)
	return args.Error(0)
}
