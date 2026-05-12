package grpc

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/nurashi/car-rental-booking/internal/api/grpc/pb"
	"github.com/nurashi/car-rental-booking/internal/domain"
	"github.com/nurashi/car-rental-booking/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements pb.BookingServiceServer.
type Server struct {
	pb.UnimplementedBookingServiceServer
	bookingSvc *service.BookingService
}

func NewServer(bookingSvc *service.BookingService) *Server {
	return &Server{bookingSvc: bookingSvc}
}

// ── 1. CreateBooking ──────────────────────────────────────────────────────────

func (s *Server) CreateBooking(ctx context.Context, req *pb.CreateBookingRequest) (*pb.BookingResponse, error) {
	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid start_date, use RFC3339 format")
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid end_date, use RFC3339 format")
	}

	input := service.CreateBookingInput{
		UserID:            req.UserId,
		VehicleID:         req.VehicleId,
		StartDate:         start,
		EndDate:           end,
		PickupLocationID:  req.PickupLocationId,
		DropoffLocationID: req.DropoffLocationId,
		Notes:             req.Notes,
	}

	b, err := s.bookingSvc.CreateBooking(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDateRange):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrVehicleNotAvailable):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		log.Printf("CreateBooking error: %v", err)
		return nil, status.Error(codes.Internal, "failed to create booking")
	}

	return &pb.BookingResponse{Booking: pb.BookingToProto(*b)}, nil
}

// ── 2. GetBooking ─────────────────────────────────────────────────────────────

func (s *Server) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.BookingResponse, error) {
	b, err := s.bookingSvc.GetBooking(ctx, req.BookingId)
	if err != nil {
		if errors.Is(err, service.ErrBookingNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		log.Printf("GetBooking error: %v", err)
		return nil, status.Error(codes.Internal, "failed to get booking")
	}
	return &pb.BookingResponse{Booking: pb.BookingToProto(*b)}, nil
}

// ── 3. UpdateBooking ──────────────────────────────────────────────────────────

func (s *Server) UpdateBooking(ctx context.Context, req *pb.UpdateBookingRequest) (*pb.BookingResponse, error) {
	input := service.UpdateBookingInput{BookingID: req.BookingId}

	if req.StartDate != nil {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid start_date")
		}
		input.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid end_date")
		}
		input.EndDate = &t
	}
	if req.Notes != nil {
		input.Notes = req.Notes
	}
	if req.PickupLocationId != nil {
		input.PickupLocationID = req.PickupLocationId
	}
	if req.DropoffLocationId != nil {
		input.DropoffLocationID = req.DropoffLocationId
	}

	b, err := s.bookingSvc.UpdateBooking(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, service.ErrInvalidDateRange):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Printf("UpdateBooking error: %v", err)
		return nil, status.Error(codes.Internal, "failed to update booking")
	}
	return &pb.BookingResponse{Booking: pb.BookingToProto(*b)}, nil
}

// ── 4. CancelBooking ──────────────────────────────────────────────────────────

func (s *Server) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	refundable, err := s.bookingSvc.CancelBooking(ctx, req.BookingId, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, service.ErrCannotCancelBooking):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		log.Printf("CancelBooking error: %v", err)
		return nil, status.Error(codes.Internal, "failed to cancel booking")
	}
	return &pb.CancelBookingResponse{
		Message:   "booking cancelled successfully",
		Refundable: refundable,
	}, nil
}

// ── 5. ListUserBookings ───────────────────────────────────────────────────────

func (s *Server) ListUserBookings(ctx context.Context, req *pb.ListUserBookingsRequest) (*pb.ListUserBookingsResponse, error) {
	bookings, total, err := s.bookingSvc.ListUserBookings(ctx, req.UserId, int(req.Page), int(req.PageSize), req.Status)
	if err != nil {
		log.Printf("ListUserBookings error: %v", err)
		return nil, status.Error(codes.Internal, "failed to list bookings")
	}
	return &pb.ListUserBookingsResponse{
		Bookings: pb.BookingsToProto(bookings),
		Total:    int32(total),
	}, nil
}

// ── 6. CalculatePrice ─────────────────────────────────────────────────────────

func (s *Server) CalculatePrice(ctx context.Context, req *pb.CalculatePriceRequest) (*pb.CalculatePriceResponse, error) {
	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid start_date")
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid end_date")
	}

	calc, err := s.bookingSvc.CalculatePrice(ctx, req.VehicleId, start, end)
	if err != nil {
		log.Printf("CalculatePrice error: %v", err)
		return nil, status.Error(codes.Internal, "failed to calculate price")
	}

	return &pb.CalculatePriceResponse{
		BasePrice:          calc.BasePrice,
		SeasonalMultiplier: calc.SeasonalMultiplier,
		TotalPrice:         calc.TotalPrice,
		RentalDays:         int32(calc.RentalDays),
		Currency:           calc.Currency,
		PricingTierId:      calc.PricingTierID,
	}, nil
}

// ── 7. CheckAvailability ──────────────────────────────────────────────────────

func (s *Server) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid start_date")
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid end_date")
	}

	result, err := s.bookingSvc.CheckAvailability(ctx, req.VehicleId, start, end)
	if err != nil {
		log.Printf("CheckAvailability error: %v", err)
		return nil, status.Error(codes.Internal, "failed to check availability")
	}

	var pbConflicts []*pb.ConflictingBooking
	for _, c := range result.Conflicts {
		pbConflicts = append(pbConflicts, &pb.ConflictingBooking{
			BookingId: c.ID,
			StartDate: c.StartDate.UTC().Format(time.RFC3339),
			EndDate:   c.EndDate.UTC().Format(time.RFC3339),
		})
	}

	return &pb.CheckAvailabilityResponse{
		Available: result.Available,
		Reason:    result.Reason,
		Conflicts: pbConflicts,
	}, nil
}

// ── 8. ExtendBooking ──────────────────────────────────────────────────────────

func (s *Server) ExtendBooking(ctx context.Context, req *pb.ExtendBookingRequest) (*pb.BookingResponse, error) {
	newEnd, err := time.Parse(time.RFC3339, req.NewEndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid new_end_date")
	}

	b, err := s.bookingSvc.ExtendBooking(ctx, req.BookingId, newEnd)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, service.ErrVehicleNotAvailable):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		log.Printf("ExtendBooking error: %v", err)
		return nil, status.Error(codes.Internal, "failed to extend booking")
	}
	return &pb.BookingResponse{Booking: pb.BookingToProto(*b)}, nil
}

// ── 9. GetActivePricingTiers ──────────────────────────────────────────────────

func (s *Server) GetActivePricingTiers(ctx context.Context, req *pb.GetActivePricingTiersRequest) (*pb.GetActivePricingTiersResponse, error) {
	tiers, err := s.bookingSvc.GetActivePricingTiers(ctx, req.VehicleCategory)
	if err != nil {
		log.Printf("GetActivePricingTiers error: %v", err)
		return nil, status.Error(codes.Internal, "failed to get pricing tiers")
	}
	return &pb.GetActivePricingTiersResponse{Tiers: pb.PricingTiersToProto(tiers)}, nil
}

// ── 10. ProcessRefund ─────────────────────────────────────────────────────────

func (s *Server) ProcessRefund(ctx context.Context, req *pb.ProcessRefundRequest) (*pb.ProcessRefundResponse, error) {
	result, err := s.bookingSvc.ProcessRefund(ctx, req.BookingId, req.RefundAmount, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		}
		log.Printf("ProcessRefund error: %v", err)
		return nil, status.Error(codes.Internal, "failed to process refund")
	}
	return &pb.ProcessRefundResponse{
		RefundId: result.RefundID,
		Amount:   result.Amount,
		Status:   result.Status,
		Message:  result.Message,
	}, nil
}

// ── 11. GetBookingHistory ─────────────────────────────────────────────────────

func (s *Server) GetBookingHistory(ctx context.Context, req *pb.GetBookingHistoryRequest) (*pb.GetBookingHistoryResponse, error) {
	bookings, total, err := s.bookingSvc.GetBookingHistory(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		log.Printf("GetBookingHistory error: %v", err)
		return nil, status.Error(codes.Internal, "failed to get booking history")
	}

	var entries []*pb.BookingHistoryEntry
	for _, b := range bookings {
		events := deriveEvents(b) //nolint:govet
		entries = append(entries, &pb.BookingHistoryEntry{
			Booking: pb.BookingToProto(b),
			Events:  events,
		})
	}

	return &pb.GetBookingHistoryResponse{Entries: entries, Total: int32(total)}, nil
}

// ── 12. ConfirmBookingPayment ─────────────────────────────────────────────────

func (s *Server) ConfirmBookingPayment(ctx context.Context, req *pb.ConfirmBookingPaymentRequest) (*pb.BookingResponse, error) {
	b, err := s.bookingSvc.ConfirmBookingPayment(ctx, req.BookingId, req.PaymentMethod, req.PaymentRef)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, service.ErrBookingAlreadyPaid):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		log.Printf("ConfirmBookingPayment error: %v", err)
		return nil, status.Error(codes.Internal, "failed to confirm payment")
	}
	return &pb.BookingResponse{Booking: pb.BookingToProto(*b)}, nil
}

// ── interface assertion ───────────────────────────────────────────────────────

var _ pb.BookingServiceServer = (*Server)(nil)

// ── helpers ───────────────────────────────────────────────────────────────────

func deriveEvents(b domain.Booking) []string {
	events := []string{"created"}
	switch b.Status {
	case domain.BookingStatusConfirmed:
		events = append(events, "confirmed")
	case domain.BookingStatusActive:
		events = append(events, "confirmed", "activated")
	case domain.BookingStatusCompleted:
		events = append(events, "confirmed", "activated", "completed")
	case domain.BookingStatusCancelled:
		events = append(events, "cancelled")
	}
	return events
}
