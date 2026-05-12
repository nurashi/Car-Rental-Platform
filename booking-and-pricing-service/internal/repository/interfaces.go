package repository

import (
	"context"

	"github.com/nurashi/car-rental-booking/internal/domain"
)

// BookingRepository defines persistence operations for bookings.
type BookingRepository interface {
	Create(ctx context.Context, b *domain.Booking) error
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	Update(ctx context.Context, b *domain.Booking) error
	Cancel(ctx context.Context, id, reason string) error
	ListByUserID(ctx context.Context, userID string, page, pageSize int, status string) ([]domain.Booking, int, error)
	CheckConflict(ctx context.Context, vehicleID, excludeBookingID string, start, end interface{}) ([]domain.Booking, error)
	ExtendBooking(ctx context.Context, id string, newEnd interface{}) (*domain.Booking, error)
	ConfirmPayment(ctx context.Context, id, paymentMethod, paymentRef string) (*domain.Booking, error)
	ListByVehicleAndDateRange(ctx context.Context, vehicleID string, start, end interface{}) ([]domain.Booking, error)
}

// PricingRepository defines persistence operations for pricing tiers.
type PricingRepository interface {
	GetTiersByCategory(ctx context.Context, vehicleCategory string) ([]domain.PricingTier, error)
	GetAllActive(ctx context.Context) ([]domain.PricingTier, error)
	GetByID(ctx context.Context, id string) (*domain.PricingTier, error)
}

// RefundRepository defines persistence operations for refunds.
type RefundRepository interface {
	Create(ctx context.Context, r *domain.Refund) error
	GetByBookingID(ctx context.Context, bookingID string) ([]domain.Refund, error)
}
