package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nurashi/car-rental-booking/internal/domain"
)

// ── Subject constants (coordinate with vehicle-inventory-service team) ────────
const (
	// Outbound – published by this service
	SubjectBookingCreated   = "booking.created"
	SubjectBookingCancelled = "booking.cancelled"

	// Inbound – published by vehicle-inventory-service
	SubjectVehicleStatusChanged = "vehicle.status.changed"
)

// ── Payloads ──────────────────────────────────────────────────────────────────

type BookingCreatedEvent struct {
	BookingID  string    `json:"booking_id"`
	UserID     string    `json:"user_id"`
	VehicleID  string    `json:"vehicle_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}

type BookingCancelledEvent struct {
	BookingID  string    `json:"booking_id"`
	UserID     string    `json:"user_id"`
	VehicleID  string    `json:"vehicle_id"`
	Reason     string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// vehicleStatusChangedPayload matches what the vehicle-inventory-service publishes.
type vehicleStatusChangedPayload struct {
	VehicleID   string    `json:"vehicle_id"`
	Status      string    `json:"status"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Description string    `json:"description"`
}

// ── EventPublisher ────────────────────────────────────────────────────────────

type EventPublisher struct {
	pub *Publisher
}

func NewEventPublisher(pub *Publisher) *EventPublisher {
	return &EventPublisher{pub: pub}
}

func (p *EventPublisher) PublishBookingCreated(b *domain.Booking) {
	event := BookingCreatedEvent{
		BookingID:  b.ID,
		UserID:     b.UserID,
		VehicleID:  b.VehicleID,
		StartDate:  b.StartDate,
		EndDate:    b.EndDate,
		TotalPrice: b.TotalPrice,
		CreatedAt:  b.CreatedAt,
	}
	if err := p.pub.Publish(SubjectBookingCreated, event); err != nil {
		log.Printf("failed to publish %s: %v", SubjectBookingCreated, err)
	}
}

func (p *EventPublisher) PublishBookingCancelled(b *domain.Booking) {
	event := BookingCancelledEvent{
		BookingID:   b.ID,
		UserID:      b.UserID,
		VehicleID:   b.VehicleID,
		Reason:      b.CancellationReason,
		CancelledAt: time.Now(),
	}
	if err := p.pub.Publish(SubjectBookingCancelled, event); err != nil {
		log.Printf("failed to publish %s: %v", SubjectBookingCancelled, err)
	}
}

// ── Inbound handlers ──────────────────────────────────────────────────────────

// MaintenanceHandler is the function called when a vehicle maintenance event arrives.
type MaintenanceHandler func(ctx context.Context, event domain.VehicleMaintenanceEvent)

func SetupInboundHandlers(sub *Subscriber, maintenanceHandler MaintenanceHandler) {
	if err := sub.Subscribe(SubjectVehicleStatusChanged, func(data []byte) {
		var payload vehicleStatusChangedPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("vehicle.status.changed: unmarshal error: %v", err)
			return
		}

		log.Printf("received %s: vehicle=%s status=%s", SubjectVehicleStatusChanged, payload.VehicleID, payload.Status)

		event := domain.VehicleMaintenanceEvent{
			VehicleID:   payload.VehicleID,
			Status:      payload.Status,
			StartDate:   payload.StartDate,
			EndDate:     payload.EndDate,
			Description: payload.Description,
		}
		maintenanceHandler(context.Background(), event)
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", SubjectVehicleStatusChanged, err)
	}
}
