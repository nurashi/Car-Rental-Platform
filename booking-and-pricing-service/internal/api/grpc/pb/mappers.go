package pb

import (
	"github.com/nurashi/car-rental-booking/internal/domain"
)

// BookingToProto converts a domain.Booking to the protobuf Booking message.
func BookingToProto(b domain.Booking) *Booking {
	return &Booking{
		Id:                 b.ID,
		UserId:             b.UserID,
		VehicleId:          b.VehicleID,
		Status:             string(b.Status),
		StartDate:          b.StartDate.UTC().Format("2006-01-02T15:04:05Z"),
		EndDate:            b.EndDate.UTC().Format("2006-01-02T15:04:05Z"),
		TotalPrice:         b.TotalPrice,
		Currency:           b.Currency,
		PickupLocationId:   b.PickupLocationID,
		DropoffLocationId:  b.DropoffLocationID,
		Notes:              b.Notes,
		CancellationReason: b.CancellationReason,
		PaymentStatus:      string(b.PaymentStatus),
		PaymentRef:         b.PaymentRef,
		CreatedAt:          b.CreatedAt.Unix(),
		UpdatedAt:          b.UpdatedAt.Unix(),
	}
}

// BookingsToProto converts a slice of domain.Booking to a slice of pb.Booking pointers.
func BookingsToProto(bookings []domain.Booking) []*Booking {
	result := make([]*Booking, len(bookings))
	for i, b := range bookings {
		result[i] = BookingToProto(b)
	}
	return result
}

// PricingTierToProto converts a domain.PricingTier to the protobuf PricingTier message.
func PricingTierToProto(t domain.PricingTier) *PricingTier {
	return &PricingTier{
		Id:                 t.ID,
		Name:               t.Name,
		VehicleCategory:    t.VehicleCategory,
		BaseDailyRate:      t.BaseDailyRate,
		SeasonalMultiplier: t.SeasonalMultiplier,
		SeasonStart:        t.SeasonStart,
		SeasonEnd:          t.SeasonEnd,
		IsActive:           t.IsActive,
		CreatedAt:          t.CreatedAt.Unix(),
		UpdatedAt:          t.UpdatedAt.Unix(),
	}
}

// PricingTiersToProto converts a slice of domain.PricingTier.
func PricingTiersToProto(tiers []domain.PricingTier) []*PricingTier {
	result := make([]*PricingTier, len(tiers))
	for i, t := range tiers {
		result[i] = PricingTierToProto(t)
	}
	return result
}
