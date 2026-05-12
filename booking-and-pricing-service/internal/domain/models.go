package domain

import "time"

// BookingStatus represents the lifecycle of a booking.
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// PaymentStatus represents the payment state of a booking.
type PaymentStatus string

const (
	PaymentStatusUnpaid            PaymentStatus = "unpaid"
	PaymentStatusPaid              PaymentStatus = "paid"
	PaymentStatusRefunded          PaymentStatus = "refunded"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
)

// Booking is the core aggregate of this service.
type Booking struct {
	ID                 string
	UserID             string
	VehicleID          string
	Status             BookingStatus
	StartDate          time.Time
	EndDate            time.Time
	TotalPrice         float64
	Currency           string
	PickupLocationID   string
	DropoffLocationID  string
	Notes              string
	CancellationReason string
	PaymentStatus      PaymentStatus
	PaymentRef         string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// RentalDays returns the number of days of the rental (minimum 1).
func (b *Booking) RentalDays() int {
	days := int(b.EndDate.Sub(b.StartDate).Hours() / 24)
	if days < 1 {
		return 1
	}
	return days
}

// PricingTier holds a single pricing configuration.
type PricingTier struct {
	ID                 string
	Name               string
	VehicleCategory    string
	BaseDailyRate      float64
	SeasonalMultiplier float64
	SeasonStart        string // "MM-DD"
	SeasonEnd          string // "MM-DD"
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Refund records a refund for a cancelled booking.
type Refund struct {
	ID        string
	BookingID string
	Amount    float64
	Reason    string
	Status    string
	CreatedAt time.Time
}

// VehicleMaintenanceEvent is used internally when NATS delivers a maintenance event.
type VehicleMaintenanceEvent struct {
	VehicleID   string    `json:"vehicle_id"`
	Status      string    `json:"status"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Description string    `json:"description"`
}
