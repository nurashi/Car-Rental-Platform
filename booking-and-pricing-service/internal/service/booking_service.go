package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nurashi/car-rental-booking/internal/domain"
	"github.com/nurashi/car-rental-booking/internal/repository"
	"github.com/redis/go-redis/v9"
)

var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrVehicleNotAvailable  = errors.New("vehicle is not available for the requested dates")
	ErrInvalidDateRange     = errors.New("end date must be after start date")
	ErrBookingAlreadyPaid   = errors.New("booking is already paid")
	ErrCannotCancelBooking  = errors.New("booking cannot be cancelled in its current state")
)

// VehicleServiceClient is the interface for calling the vehicle inventory service.
type VehicleServiceClient interface {
	GetVehicleCategory(ctx context.Context, vehicleID string) (string, error)
	IsVehicleAvailable(ctx context.Context, vehicleID string) (bool, error)
	CheckVehicleAvailability(ctx context.Context, vehicleID string) (bool, float64)
}

// NoopVehicleClient is used when the vehicle service is unreachable (graceful degradation).
type NoopVehicleClient struct{}

func (n *NoopVehicleClient) GetVehicleCategory(_ context.Context, _ string) (string, error) {
	return "economy", nil
}
func (n *NoopVehicleClient) IsVehicleAvailable(_ context.Context, _ string) (bool, error) {
	return true, nil
}
func (n *NoopVehicleClient) CheckVehicleAvailability(_ context.Context, _ string) (bool, float64) {
	return true, 50.0
}

// BookingService contains all booking and pricing business logic.
type BookingService struct {
	bookingRepo repository.BookingRepository
	pricingRepo repository.PricingRepository
	refundRepo  repository.RefundRepository
	redisClient *redis.Client
	vehicleSvc  VehicleServiceClient
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	pricingRepo repository.PricingRepository,
	refundRepo repository.RefundRepository,
	redisClient *redis.Client,
	vehicleSvc VehicleServiceClient,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		pricingRepo: pricingRepo,
		refundRepo:  refundRepo,
		redisClient: redisClient,
		vehicleSvc:  vehicleSvc,
	}
}

// ── Booking operations ───────────────────────────────────────────────────────

type CreateBookingInput struct {
	UserID            string
	VehicleID         string
	StartDate         time.Time
	EndDate           time.Time
	PickupLocationID  string
	DropoffLocationID string
	Notes             string
}

func (s *BookingService) CreateBooking(ctx context.Context, input CreateBookingInput) (*domain.Booking, error) {
	if !input.EndDate.After(input.StartDate) {
		return nil, ErrInvalidDateRange
	}

	// Check vehicle status via inventory service
	available, err := s.vehicleSvc.IsVehicleAvailable(ctx, input.VehicleID)
	if err != nil {
		log.Printf("vehicle availability check failed: %v", err)
		// degrade gracefully – do not block booking if vehicle service is down
	} else if !available {
		return nil, ErrVehicleNotAvailable
	}

	// Fetch vehicle category for pricing
	category, err := s.vehicleSvc.GetVehicleCategory(ctx, input.VehicleID)
	if err != nil {
		log.Printf("vehicle category fetch failed, using default pricing: %v", err)
		category = "economy"
	}

	// Calculate price (cache-aside)
	price, _, err := s.calculatePrice(ctx, input.VehicleID, category, input.StartDate, input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("calculate price: %w", err)
	}

	b := &domain.Booking{
		ID:                uuid.New().String(),
		UserID:            input.UserID,
		VehicleID:         input.VehicleID,
		Status:            domain.BookingStatusPending,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		TotalPrice:        price,
		Currency:          "USD",
		PickupLocationID:  input.PickupLocationID,
		DropoffLocationID: input.DropoffLocationID,
		Notes:             input.Notes,
		PaymentStatus:     domain.PaymentStatusUnpaid,
	}

	// Conflict check + insert happen inside a DB transaction (called by repo)
	if err := s.bookingRepo.Create(ctx, b); err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}

	return b, nil
}

func (s *BookingService) GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error) {
	b, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking: %w", err)
	}
	return b, nil
}

type UpdateBookingInput struct {
	BookingID         string
	StartDate         *time.Time
	EndDate           *time.Time
	Notes             *string
	PickupLocationID  *string
	DropoffLocationID *string
}

func (s *BookingService) UpdateBooking(ctx context.Context, input UpdateBookingInput) (*domain.Booking, error) {
	b, err := s.bookingRepo.GetByID(ctx, input.BookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking for update: %w", err)
	}

	if input.StartDate != nil {
		b.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		b.EndDate = *input.EndDate
	}
	if input.Notes != nil {
		b.Notes = *input.Notes
	}
	if input.PickupLocationID != nil {
		b.PickupLocationID = *input.PickupLocationID
	}
	if input.DropoffLocationID != nil {
		b.DropoffLocationID = *input.DropoffLocationID
	}

	if !b.EndDate.After(b.StartDate) {
		return nil, ErrInvalidDateRange
	}

	if err := s.bookingRepo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("update booking: %w", err)
	}
	return b, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID, reason string) (bool, error) {
	b, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return false, ErrBookingNotFound
		}
		return false, fmt.Errorf("get booking for cancel: %w", err)
	}

	if b.Status == domain.BookingStatusCancelled || b.Status == domain.BookingStatusCompleted {
		return false, ErrCannotCancelBooking
	}

	if err := s.bookingRepo.Cancel(ctx, bookingID, reason); err != nil {
		return false, fmt.Errorf("cancel booking: %w", err)
	}

	// A booking is refundable if cancelled >= 24h before start
	refundable := time.Until(b.StartDate) >= 24*time.Hour && b.PaymentStatus == domain.PaymentStatusPaid
	return refundable, nil
}

func (s *BookingService) ListUserBookings(ctx context.Context, userID string, page, pageSize int, status string) ([]domain.Booking, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.bookingRepo.ListByUserID(ctx, userID, page, pageSize, status)
}

func (s *BookingService) ExtendBooking(ctx context.Context, bookingID string, newEnd time.Time) (*domain.Booking, error) {
	b, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking for extend: %w", err)
	}

	if !newEnd.After(b.EndDate) {
		return nil, errors.New("new end date must be after current end date")
	}

	// Check no conflict with the extended date range
	conflicts, err := s.bookingRepo.CheckConflict(ctx, b.VehicleID, b.ID, b.StartDate, newEnd)
	if err != nil {
		return nil, fmt.Errorf("conflict check for extend: %w", err)
	}
	if len(conflicts) > 0 {
		return nil, ErrVehicleNotAvailable
	}

	extended, err := s.bookingRepo.ExtendBooking(ctx, bookingID, newEnd)
	if err != nil {
		return nil, fmt.Errorf("extend booking: %w", err)
	}
	return extended, nil
}

func (s *BookingService) ConfirmBookingPayment(ctx context.Context, bookingID, paymentMethod, paymentRef string) (*domain.Booking, error) {
	b, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking for payment: %w", err)
	}

	if b.PaymentStatus == domain.PaymentStatusPaid {
		return nil, ErrBookingAlreadyPaid
	}

	confirmed, err := s.bookingRepo.ConfirmPayment(ctx, bookingID, paymentMethod, paymentRef)
	if err != nil {
		return nil, fmt.Errorf("confirm payment: %w", err)
	}
	return confirmed, nil
}

// ── Availability & pricing ───────────────────────────────────────────────────

type AvailabilityResult struct {
	Available bool            `json:"available"`
	Reason    string          `json:"reason"`
	Conflicts []domain.Booking `json:"conflicts"`
}

func (s *BookingService) CheckAvailability(ctx context.Context, vehicleID string, start, end time.Time) (*AvailabilityResult, error) {
	if !end.After(start) {
		return &AvailabilityResult{Available: false, Reason: "invalid date range"}, nil
	}

	conflicts, err := s.bookingRepo.ListByVehicleAndDateRange(ctx, vehicleID, start, end)
	if err != nil {
		return nil, fmt.Errorf("check availability: %w", err)
	}

	if len(conflicts) > 0 {
		return &AvailabilityResult{
			Available: false,
			Reason:    "vehicle already booked for requested dates",
			Conflicts: conflicts,
		}, nil
	}

	// Also check vehicle service status
	available, err := s.vehicleSvc.IsVehicleAvailable(ctx, vehicleID)
	if err != nil {
		log.Printf("vehicle service unavailable, skipping status check: %v", err)
	} else if !available {
		return &AvailabilityResult{Available: false, Reason: "vehicle is under maintenance or unavailable"}, nil
	}

	return &AvailabilityResult{Available: true}, nil
}

type PriceCalculation struct {
	BasePrice          float64 `json:"base_price"`
	SeasonalMultiplier float64 `json:"seasonal_multiplier"`
	TotalPrice         float64 `json:"total_price"`
	RentalDays         int     `json:"rental_days"`
	Currency           string  `json:"currency"`
	PricingTierID      string  `json:"pricing_tier_id"`
}

func (s *BookingService) CalculatePrice(ctx context.Context, vehicleID string, start, end time.Time) (*PriceCalculation, error) {
	category, err := s.vehicleSvc.GetVehicleCategory(ctx, vehicleID)
	if err != nil {
		log.Printf("vehicle category fetch failed, using default: %v", err)
		category = "economy"
	}

	price, tier, err := s.calculatePrice(ctx, vehicleID, category, start, end)
	if err != nil {
		return nil, err
	}

	days := int(end.Sub(start).Hours() / 24)
	if days < 1 {
		days = 1
	}

	multiplier := 1.0
	if tier != nil {
		multiplier = tier.SeasonalMultiplier
	}

	return &PriceCalculation{
		BasePrice:          price / multiplier,
		SeasonalMultiplier: multiplier,
		TotalPrice:         price,
		RentalDays:         days,
		Currency:           "USD",
		PricingTierID:      func() string { if tier != nil { return tier.ID }; return "" }(),
	}, nil
}

// calculatePrice is the internal cache-aside pricing engine.
func (s *BookingService) calculatePrice(ctx context.Context, vehicleID, category string, start, end time.Time) (float64, *domain.PricingTier, error) {
	days := int(end.Sub(start).Hours() / 24)
	if days < 1 {
		days = 1
	}

	// Try Redis cache first
	cacheKey := fmt.Sprintf("pricing:category:%s", category)
	var tiers []domain.PricingTier

	if s.redisClient != nil {
		cached, cacheErr := s.redisClient.Get(ctx, cacheKey).Bytes()
		if cacheErr == nil {
			if jsonErr := json.Unmarshal(cached, &tiers); jsonErr != nil {
				log.Printf("pricing cache unmarshal error: %v", jsonErr)
			}
		}
	}

	// Cache miss – load from DB and store
	if len(tiers) == 0 {
		var err error
		tiers, err = s.pricingRepo.GetTiersByCategory(ctx, category)
		if err != nil {
			return 0, nil, fmt.Errorf("get pricing tiers: %w", err)
		}
		if len(tiers) == 0 {
			// Fallback: get any economy tier
			tiers, _ = s.pricingRepo.GetTiersByCategory(ctx, "economy")
		}

		if payload, jsonErr := json.Marshal(tiers); jsonErr == nil && s.redisClient != nil {
			s.redisClient.Set(ctx, cacheKey, payload, 10*time.Minute)
		}
	}

	if len(tiers) == 0 {
		return float64(days) * 30.0, nil, nil // hard fallback
	}

	// Pick the best matching tier (seasonal > base)
	startMD := start.Format("01-02")
	var best *domain.PricingTier
	for i := range tiers {
		t := &tiers[i]
		if t.SeasonStart != "" && t.SeasonEnd != "" {
			if startMD >= t.SeasonStart && startMD <= t.SeasonEnd {
				best = t
				break
			}
		}
	}
	if best == nil {
		// Use the standard tier (lowest multiplier)
		for i := range tiers {
			if tiers[i].SeasonStart == "" {
				best = &tiers[i]
				break
			}
		}
	}
	if best == nil {
		best = &tiers[0]
	}

	total := best.BaseDailyRate * best.SeasonalMultiplier * float64(days)
	return total, best, nil
}

// ── Pricing tiers ────────────────────────────────────────────────────────────

func (s *BookingService) GetActivePricingTiers(ctx context.Context, vehicleCategory string) ([]domain.PricingTier, error) {
	cacheKey := fmt.Sprintf("pricing:active:%s", vehicleCategory)
	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var tiers []domain.PricingTier
			if jsonErr := json.Unmarshal(cached, &tiers); jsonErr == nil {
				return tiers, nil
			}
		}
	}

	var tiers []domain.PricingTier
	var err error
	if vehicleCategory != "" {
		tiers, err = s.pricingRepo.GetTiersByCategory(ctx, vehicleCategory)
	} else {
		tiers, err = s.pricingRepo.GetAllActive(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("get active pricing tiers: %w", err)
	}

	if payload, jsonErr := json.Marshal(tiers); jsonErr == nil && s.redisClient != nil {
		s.redisClient.Set(ctx, cacheKey, payload, 10*time.Minute)
	}

	return tiers, nil
}

// ── Refunds ──────────────────────────────────────────────────────────────────

type RefundResult struct {
	RefundID string
	Amount   float64
	Status   string
	Message  string
}

func (s *BookingService) ProcessRefund(ctx context.Context, bookingID string, refundAmount float64, reason string) (*RefundResult, error) {
	b, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking for refund: %w", err)
	}

	if b.PaymentStatus != domain.PaymentStatusPaid {
		return nil, errors.New("booking is not paid, no refund applicable")
	}

	if refundAmount <= 0 || refundAmount > b.TotalPrice {
		refundAmount = b.TotalPrice
	}

	refund := &domain.Refund{
		ID:        uuid.New().String(),
		BookingID: bookingID,
		Amount:    refundAmount,
		Reason:    reason,
		Status:    "processed",
	}
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, fmt.Errorf("create refund: %w", err)
	}

	return &RefundResult{
		RefundID: refund.ID,
		Amount:   refund.Amount,
		Status:   refund.Status,
		Message:  "refund processed successfully",
	}, nil
}

// ── History ──────────────────────────────────────────────────────────────────

func (s *BookingService) GetBookingHistory(ctx context.Context, userID string, page, pageSize int) ([]domain.Booking, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	// History = all bookings (including completed and cancelled)
	return s.bookingRepo.ListByUserID(ctx, userID, page, pageSize, "")
}

// ── NATS event handler ───────────────────────────────────────────────────────

// HandleVehicleMaintenanceEvent processes incoming maintenance events from NATS.
// It cancels any confirmed/pending bookings that conflict with the maintenance window.
func (s *BookingService) HandleVehicleMaintenanceEvent(ctx context.Context, event domain.VehicleMaintenanceEvent) {
	if event.Status != "maintenance" {
		return
	}

	conflicts, err := s.bookingRepo.ListByVehicleAndDateRange(ctx, event.VehicleID, event.StartDate, event.EndDate)
	if err != nil {
		log.Printf("maintenance handler: list conflicts error: %v", err)
		return
	}

	for _, b := range conflicts {
		reason := fmt.Sprintf("vehicle scheduled for maintenance: %s", event.Description)
		if err := s.bookingRepo.Cancel(ctx, b.ID, reason); err != nil {
			log.Printf("maintenance handler: cancel booking %s error: %v", b.ID, err)
		} else {
			log.Printf("maintenance handler: cancelled booking %s due to vehicle maintenance", b.ID)
		}
	}
}
