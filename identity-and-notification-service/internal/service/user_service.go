package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/repository"
)

var ErrProfileNotFound = errors.New("profile not found")

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type UpdateProfileInput struct {
	UserID    string
	FirstName *string
	LastName  *string
	Phone     *string
	AvatarURL *string
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.Phone != nil {
		user.Phone = *input.Phone
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userRepo.List(ctx, page, pageSize)
}

func (s *UserService) BlockUser(ctx context.Context, userID, reason string) error {
	if err := s.userRepo.BlockUser(ctx, userID, reason); err != nil {
		return fmt.Errorf("block user: %w", err)
	}
	return nil
}
