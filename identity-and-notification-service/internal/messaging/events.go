package messaging

import (
	"encoding/json"
	"log"
)

const (
	EventUserRegistered     = "user.registered"
	EventEmailVerified      = "email.verified"
	EventProfileUpdated     = "user.profile.updated"
	EventUserBlocked        = "user.blocked"
	EventLicenseSubmitted   = "license.submitted"
	EventLicenseValidated   = "license.validated"

	// Inbound events from other services
	EventBookingCreated = "booking.created"
	EventVehicleStatusChanged = "vehicle.status.changed"
)

type EventPublisher struct {
	pub *Publisher
}

func NewEventPublisher(pub *Publisher) *EventPublisher {
	return &EventPublisher{pub: pub}
}

func (p *EventPublisher) PublishUserRegistered(userID, email string) {
	event := map[string]interface{}{
		"user_id": userID,
		"email":   email,
	}
	if err := p.pub.Publish(EventUserRegistered, event); err != nil {
		log.Printf("failed to publish %s: %v", EventUserRegistered, err)
	}
}

func (p *EventPublisher) PublishEmailVerified(userID, email string) {
	event := map[string]interface{}{
		"user_id": userID,
		"email":   email,
	}
	if err := p.pub.Publish(EventEmailVerified, event); err != nil {
		log.Printf("failed to publish %s: %v", EventEmailVerified, err)
	}
}

func (p *EventPublisher) PublishProfileUpdated(userID string) {
	event := map[string]interface{}{
		"user_id": userID,
	}
	if err := p.pub.Publish(EventProfileUpdated, event); err != nil {
		log.Printf("failed to publish %s: %v", EventProfileUpdated, err)
	}
}

func (p *EventPublisher) PublishUserBlocked(userID, reason string) {
	event := map[string]interface{}{
		"user_id": userID,
		"reason":  reason,
	}
	if err := p.pub.Publish(EventUserBlocked, event); err != nil {
		log.Printf("failed to publish %s: %v", EventUserBlocked, err)
	}
}

func (p *EventPublisher) PublishLicenseSubmitted(userID, licenseID string) {
	event := map[string]interface{}{
		"user_id":    userID,
		"license_id": licenseID,
	}
	if err := p.pub.Publish(EventLicenseSubmitted, event); err != nil {
		log.Printf("failed to publish %s: %v", EventLicenseSubmitted, err)
	}
}

func (p *EventPublisher) PublishLicenseValidated(userID string, isValid bool) {
	event := map[string]interface{}{
		"user_id":  userID,
		"is_valid": isValid,
	}
	if err := p.pub.Publish(EventLicenseValidated, event); err != nil {
		log.Printf("failed to publish %s: %v", EventLicenseValidated, err)
	}
}

type EventHandler func(event []byte)

func (h EventHandler) Handle(data []byte) {
	h(data)
}

func SetupInboundHandlers(sub *Subscriber, handler func(subject string, data []byte)) {
	subscribe := func(subject string) {
		sub.Subscribe(subject, func(data []byte) {
			var event map[string]interface{}
			if err := json.Unmarshal(data, &event); err != nil {
				log.Printf("failed to unmarshal event on %s: %v", subject, err)
				return
			}
			log.Printf("received event %s: %v", subject, event)
			handler(subject, data)
		})
	}

	subscribe(EventBookingCreated)
	subscribe(EventVehicleStatusChanged)
}
