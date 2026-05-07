package repository

import (
	"context"

	"github.com/nurashi/car-rental-identity/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateEmailVerified(ctx context.Context, userID string, verified bool) error
	List(ctx context.Context, page, pageSize int) ([]domain.User, int, error)
	BlockUser(ctx context.Context, userID, reason string) error
}

type LicenseRepository interface {
	Create(ctx context.Context, license *domain.DriverLicense) error
	GetByUserID(ctx context.Context, userID string) (*domain.DriverLicense, error)
	Update(ctx context.Context, license *domain.DriverLicense) error
}

type EmailVerificationRepository interface {
	Create(ctx context.Context, ev *domain.EmailVerification) error
	GetByToken(ctx context.Context, token string) (*domain.EmailVerification, error)
	MarkUsed(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.NotificationRecord) error
	GetByUserID(ctx context.Context, userID string, limit int) ([]domain.NotificationRecord, error)
}
