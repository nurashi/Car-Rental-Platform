package service_test

import (
	"context"
	"testing"

	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/repository"
	"github.com/nurashi/car-rental-identity/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_GetProfile(t *testing.T) {
	userRepo := new(mockUserRepo)
	expectedUser := &domain.User{ID: "1", Email: "test@example.com", FirstName: "John"}
	userRepo.On("GetByID", mock.Anything, "1").Return(expectedUser, nil)

	userService := service.NewUserService(userRepo)
	user, err := userService.GetProfile(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, "John", user.FirstName)
	userRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_NotFound(t *testing.T) {
	userRepo := new(mockUserRepo)
	userRepo.On("GetByID", mock.Anything, "nonexistent").Return((*domain.User)(nil), repository.ErrUserNotFound)

	userService := service.NewUserService(userRepo)
	user, err := userService.GetProfile(context.Background(), "nonexistent")

	assert.Nil(t, user)
	assert.Equal(t, service.ErrProfileNotFound, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile(t *testing.T) {
	userRepo := new(mockUserRepo)
	existingUser := &domain.User{ID: "1", Email: "test@example.com", FirstName: "John", LastName: "Doe"}
	userRepo.On("GetByID", mock.Anything, "1").Return(existingUser, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	userService := service.NewUserService(userRepo)
	newName := "Jane"
	input := service.UpdateProfileInput{UserID: "1", FirstName: &newName}

	updated, err := userService.UpdateProfile(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, "Jane", updated.FirstName)
	userRepo.AssertExpectations(t)
}

func TestUserService_BlockUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	userRepo.On("BlockUser", mock.Anything, "1", "violation").Return(nil)

	userService := service.NewUserService(userRepo)
	err := userService.BlockUser(context.Background(), "1", "violation")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_ListUsers(t *testing.T) {
	userRepo := new(mockUserRepo)
	users := []domain.User{{ID: "1", Email: "test@example.com"}}
	userRepo.On("List", mock.Anything, 1, 20).Return(users, 1, nil)

	userService := service.NewUserService(userRepo)
	result, total, err := userService.ListUsers(context.Background(), 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
	userRepo.AssertExpectations(t)
}
