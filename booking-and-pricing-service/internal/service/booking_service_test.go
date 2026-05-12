package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/nurashi/car-rental-booking/internal/domain"
	"github.com/nurashi/car-rental-booking/internal/repository"
	"github.com/nurashi/car-rental-booking/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBookingRepo struct {
	mock.Mock
}

func (m *mockBookingRepo) Create(ctx context.Context, b *domain.Booking) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *mockBookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) Update(ctx context.Context, b *domain.Booking) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *mockBookingRepo) Cancel(ctx context.Context, id, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *mockBookingRepo) ListByUserID(ctx context.Context, userID string, page, pageSize int, status string) ([]domain.Booking, int, error) {
	args := m.Called(ctx, userID, page, pageSize, status)
	return args.Get(0).([]domain.Booking), args.Int(1), args.Error(2)
}

func (m *mockBookingRepo) CheckConflict(ctx context.Context, vehicleID, excludeBookingID string, start, end interface{}) ([]domain.Booking, error) {
	args := m.Called(ctx, vehicleID, excludeBookingID, start, end)
	return args.Get(0).([]domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) ExtendBooking(ctx context.Context, id string, newEnd interface{}) (*domain.Booking, error) {
	args := m.Called(ctx, id, newEnd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) ConfirmPayment(ctx context.Context, id, paymentMethod, paymentRef string) (*domain.Booking, error) {
	args := m.Called(ctx, id, paymentMethod, paymentRef)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *mockBookingRepo) ListByVehicleAndDateRange(ctx context.Context, vehicleID string, start, end interface{}) ([]domain.Booking, error) {
	args := m.Called(ctx, vehicleID, start, end)
	return args.Get(0).([]domain.Booking), args.Error(1)
}

type mockPricingRepo struct {
	mock.Mock
}

func (m *mockPricingRepo) GetTiersByCategory(ctx context.Context, category string) ([]domain.PricingTier, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]domain.PricingTier), args.Error(1)
}

func (m *mockPricingRepo) GetAllActive(ctx context.Context) ([]domain.PricingTier, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.PricingTier), args.Error(1)
}

func (m *mockPricingRepo) GetByID(ctx context.Context, id string) (*domain.PricingTier, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PricingTier), args.Error(1)
}

type mockRefundRepo struct {
	mock.Mock
}

func (m *mockRefundRepo) Create(ctx context.Context, r *domain.Refund) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *mockRefundRepo) GetByBookingID(ctx context.Context, bookingID string) ([]domain.Refund, error) {
	args := m.Called(ctx, bookingID)
	return args.Get(0).([]domain.Refund), args.Error(1)
}

type mockVehicleClient struct {
	mock.Mock
}

func (m *mockVehicleClient) GetVehicleCategory(ctx context.Context, vehicleID string) (string, error) {
	args := m.Called(ctx, vehicleID)
	return args.String(0), args.Error(1)
}

func (m *mockVehicleClient) IsVehicleAvailable(ctx context.Context, vehicleID string) (bool, error) {
	args := m.Called(ctx, vehicleID)
	return args.Bool(0), args.Error(1)
}

func (m *mockVehicleClient) CheckVehicleAvailability(ctx context.Context, vehicleID string) (bool, float64) {
	args := m.Called(ctx, vehicleID)
	return args.Bool(0), args.Get(1).(float64)
}

func TestBookingService_CreateBooking_InvalidDateRange(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)

	input := service.CreateBookingInput{
		UserID:    "u1",
		VehicleID: "v1",
		StartDate: time.Now().Add(48 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}

	result, err := svc.CreateBooking(context.Background(), input)

	assert.Nil(t, result)
	assert.Equal(t, service.ErrInvalidDateRange, err)
}

func TestBookingService_CreateBooking_VehicleNotAvailable(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	start := time.Now().Add(24 * time.Hour)
	end := time.Now().Add(48 * time.Hour)

	vehicleClient.On("IsVehicleAvailable", mock.Anything, "v1").Return(false, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)

	input := service.CreateBookingInput{
		UserID:    "u1",
		VehicleID: "v1",
		StartDate: start,
		EndDate:   end,
	}

	result, err := svc.CreateBooking(context.Background(), input)

	assert.Nil(t, result)
	assert.Equal(t, service.ErrVehicleNotAvailable, err)
}

func TestBookingService_CreateBooking_Success(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	start := time.Now().Add(24 * time.Hour)
	end := time.Now().Add(72 * time.Hour)

	vehicleClient.On("IsVehicleAvailable", mock.Anything, "v1").Return(true, nil)
	vehicleClient.On("GetVehicleCategory", mock.Anything, "v1").Return("economy", nil)
	pricingRepo.On("GetTiersByCategory", mock.Anything, "economy").Return([]domain.PricingTier{
		{ID: "t1", Name: "Standard", VehicleCategory: "economy", BaseDailyRate: 50.0, SeasonalMultiplier: 1.0},
	}, nil)
	bookingRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Booking")).Run(func(args mock.Arguments) {
		b := args.Get(1).(*domain.Booking)
		b.ID = "b1"
	}).Return(nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)

	input := service.CreateBookingInput{
		UserID:    "u1",
		VehicleID: "v1",
		StartDate: start,
		EndDate:   end,
	}

	result, err := svc.CreateBooking(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, "b1", result.ID)
	assert.Equal(t, domain.BookingStatusPending, result.Status)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_GetBooking_NotFound(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	bookingRepo.On("GetByID", mock.Anything, "nonexistent").Return((*domain.Booking)(nil), repository.ErrBookingNotFound)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, err := svc.GetBooking(context.Background(), "nonexistent")

	assert.Nil(t, result)
	assert.Equal(t, service.ErrBookingNotFound, err)
}

func TestBookingService_GetBooking_Success(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	booking := &domain.Booking{ID: "b1", UserID: "u1", VehicleID: "v1"}
	bookingRepo.On("GetByID", mock.Anything, "b1").Return(booking, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, err := svc.GetBooking(context.Background(), "b1")

	assert.NoError(t, err)
	assert.Equal(t, "b1", result.ID)
}

func TestBookingService_CancelBooking_NotFound(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	bookingRepo.On("GetByID", mock.Anything, "b1").Return((*domain.Booking)(nil), repository.ErrBookingNotFound)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	refundable, err := svc.CancelBooking(context.Background(), "b1", "changed plans")

	assert.False(t, refundable)
	assert.Equal(t, service.ErrBookingNotFound, err)
}

func TestBookingService_CancelBooking_AlreadyCancelled(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	booking := &domain.Booking{
		ID:            "b1",
		Status:        domain.BookingStatusCancelled,
		PaymentStatus: domain.PaymentStatusPaid,
		StartDate:     time.Now().Add(48 * time.Hour),
	}
	bookingRepo.On("GetByID", mock.Anything, "b1").Return(booking, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	refundable, err := svc.CancelBooking(context.Background(), "b1", "changed plans")

	assert.False(t, refundable)
	assert.Equal(t, service.ErrCannotCancelBooking, err)
}

func TestBookingService_CancelBooking_Refundable(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	booking := &domain.Booking{
		ID:            "b1",
		Status:        domain.BookingStatusConfirmed,
		PaymentStatus: domain.PaymentStatusPaid,
		StartDate:     time.Now().Add(48 * time.Hour),
	}
	bookingRepo.On("GetByID", mock.Anything, "b1").Return(booking, nil)
	bookingRepo.On("Cancel", mock.Anything, "b1", "changed plans").Return(nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	refundable, err := svc.CancelBooking(context.Background(), "b1", "changed plans")

	assert.True(t, refundable)
	assert.NoError(t, err)
	bookingRepo.AssertExpectations(t)
}

func TestBookingService_CalculatePrice(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	start := time.Now().Add(24 * time.Hour)
	end := time.Now().Add(72 * time.Hour)

	vehicleClient.On("GetVehicleCategory", mock.Anything, "v1").Return("economy", nil)
	pricingRepo.On("GetTiersByCategory", mock.Anything, "economy").Return([]domain.PricingTier{
		{ID: "t1", BaseDailyRate: 50.0, SeasonalMultiplier: 1.2},
	}, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, err := svc.CalculatePrice(context.Background(), "v1", start, end)

	assert.NoError(t, err)
	assert.Equal(t, "USD", result.Currency)
	assert.Greater(t, result.TotalPrice, float64(0))
}

func TestBookingService_CheckAvailability_Available(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	start := time.Now().Add(24 * time.Hour)
	end := time.Now().Add(48 * time.Hour)

	bookingRepo.On("ListByVehicleAndDateRange", mock.Anything, "v1", start, end).Return([]domain.Booking{}, nil)
	vehicleClient.On("IsVehicleAvailable", mock.Anything, "v1").Return(true, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, err := svc.CheckAvailability(context.Background(), "v1", start, end)

	assert.NoError(t, err)
	assert.True(t, result.Available)
}

func TestBookingService_CheckAvailability_Conflict(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	start := time.Now().Add(24 * time.Hour)
	end := time.Now().Add(48 * time.Hour)

	bookingRepo.On("ListByVehicleAndDateRange", mock.Anything, "v1", start, end).Return([]domain.Booking{{ID: "b1"}}, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, err := svc.CheckAvailability(context.Background(), "v1", start, end)

	assert.NoError(t, err)
	assert.False(t, result.Available)
	assert.Equal(t, "vehicle already booked for requested dates", result.Reason)
}

func TestBookingService_ProcessRefund_NotPaid(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	booking := &domain.Booking{ID: "b1", PaymentStatus: domain.PaymentStatusUnpaid, TotalPrice: 100.0}
	bookingRepo.On("GetByID", mock.Anything, "b1").Return(booking, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, err := svc.ProcessRefund(context.Background(), "b1", 100.0, "cancelled")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not paid")
}

func TestBookingService_ProcessRefund_Success(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	booking := &domain.Booking{ID: "b1", PaymentStatus: domain.PaymentStatusPaid, TotalPrice: 100.0}
	bookingRepo.On("GetByID", mock.Anything, "b1").Return(booking, nil)
	refundRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Refund")).Return(nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, err := svc.ProcessRefund(context.Background(), "b1", 100.0, "cancelled")

	assert.NoError(t, err)
	assert.Equal(t, "processed", result.Status)
	refundRepo.AssertExpectations(t)
}

func TestBookingService_ListUserBookings(t *testing.T) {
	bookingRepo := new(mockBookingRepo)
	pricingRepo := new(mockPricingRepo)
	refundRepo := new(mockRefundRepo)
	vehicleClient := new(mockVehicleClient)

	bookings := []domain.Booking{{ID: "b1", UserID: "u1"}}
	bookingRepo.On("ListByUserID", mock.Anything, "u1", 1, 20, "").Return(bookings, 1, nil)

	svc := service.NewBookingService(bookingRepo, pricingRepo, refundRepo, nil, vehicleClient)
	result, total, err := svc.ListUserBookings(context.Background(), "u1", 1, 20, "")

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}
