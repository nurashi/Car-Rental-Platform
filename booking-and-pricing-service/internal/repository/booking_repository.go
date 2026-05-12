package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/car-rental-booking/internal/domain"
)

var ErrBookingNotFound = errors.New("booking not found")

type BookingRepo struct {
	pool *pgxpool.Pool
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func NewBookingRepo(pool *pgxpool.Pool) *BookingRepo {
	return &BookingRepo{pool: pool}
}

func (r *BookingRepo) Create(ctx context.Context, b *domain.Booking) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	if b.Currency == "" {
		b.Currency = "USD"
	}

	query := `
		INSERT INTO bookings (
			id, user_id, vehicle_id, status, start_date, end_date,
			total_price, currency, pickup_location_id, dropoff_location_id,
			notes, payment_status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`
	_, err := r.pool.Exec(ctx, query,
		b.ID, b.UserID, b.VehicleID, string(b.Status),
		b.StartDate, b.EndDate, b.TotalPrice, b.Currency,
		nullString(b.PickupLocationID), nullString(b.DropoffLocationID),
		nullString(b.Notes), string(b.PaymentStatus),
		b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create booking: %w", err)
	}
	return nil
}

func (r *BookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	query := `
		SELECT id, user_id, vehicle_id, status, start_date, end_date,
		       total_price, currency,
		       COALESCE(pickup_location_id::text, ''),
		       COALESCE(dropoff_location_id::text, ''),
		       COALESCE(notes, ''),
		       COALESCE(cancellation_reason, ''),
		       payment_status,
		       COALESCE(payment_ref, ''),
		       created_at, updated_at
		FROM bookings WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanBooking(row)
}

func (r *BookingRepo) Update(ctx context.Context, b *domain.Booking) error {
	b.UpdatedAt = time.Now()
	query := `
		UPDATE bookings
		SET start_date=$2, end_date=$3, notes=$4,
		    pickup_location_id=$5, dropoff_location_id=$6,
		    total_price=$7, updated_at=$8
		WHERE id=$1
	`
	result, err := r.pool.Exec(ctx, query,
		b.ID, b.StartDate, b.EndDate, nullString(b.Notes),
		nullString(b.PickupLocationID), nullString(b.DropoffLocationID),
		b.TotalPrice, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update booking: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepo) Cancel(ctx context.Context, id, reason string) error {
	query := `
		UPDATE bookings
		SET status='cancelled', cancellation_reason=$2, updated_at=$3
		WHERE id=$1 AND status NOT IN ('cancelled','completed')
	`
	result, err := r.pool.Exec(ctx, query, id, reason, time.Now())
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepo) ListByUserID(ctx context.Context, userID string, page, pageSize int, status string) ([]domain.Booking, int, error) {
	offset := (page - 1) * pageSize
	if page < 1 {
		offset = 0
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM bookings WHERE user_id = $1`
	countArgs := []interface{}{userID}

	if status != "" {
		countQuery += ` AND status = $2`
		countArgs = append(countArgs, status)
	}
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count bookings: %w", err)
	}

	listQuery := `
		SELECT id, user_id, vehicle_id, status, start_date, end_date,
		       total_price, currency,
		       COALESCE(pickup_location_id::text, ''),
		       COALESCE(dropoff_location_id::text, ''),
		       COALESCE(notes, ''),
		       COALESCE(cancellation_reason, ''),
		       payment_status,
		       COALESCE(payment_ref, ''),
		       created_at, updated_at
		FROM bookings WHERE user_id = $1
	`
	listArgs := []interface{}{userID}
	if status != "" {
		listQuery += ` AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		listArgs = append(listArgs, status, pageSize, offset)
	} else {
		listQuery += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		listArgs = append(listArgs, pageSize, offset)
	}

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBookingRow(rows)
		if err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, *b)
	}
	return bookings, total, nil
}

// CheckConflict checks for overlapping bookings for the same vehicle using SELECT FOR UPDATE.
func (r *BookingRepo) CheckConflict(ctx context.Context, vehicleID, excludeBookingID string, start, end interface{}) ([]domain.Booking, error) {
	query := `
		SELECT id, user_id, vehicle_id, status, start_date, end_date,
		       total_price, currency,
		       COALESCE(pickup_location_id::text, ''),
		       COALESCE(dropoff_location_id::text, ''),
		       COALESCE(notes, ''),
		       COALESCE(cancellation_reason, ''),
		       payment_status,
		       COALESCE(payment_ref, ''),
		       created_at, updated_at
		FROM bookings
		WHERE vehicle_id = $1
		  AND status NOT IN ('cancelled', 'completed')
		  AND id != $2
		  AND start_date < $4
		  AND end_date   > $3
		FOR UPDATE
	`
	rows, err := r.pool.Query(ctx, query, vehicleID, excludeBookingID, start, end)
	if err != nil {
		return nil, fmt.Errorf("check conflict: %w", err)
	}
	defer rows.Close()

	var conflicts []domain.Booking
	for rows.Next() {
		b, err := scanBookingRow(rows)
		if err != nil {
			return nil, err
		}
		conflicts = append(conflicts, *b)
	}
	return conflicts, nil
}

func (r *BookingRepo) ExtendBooking(ctx context.Context, id string, newEnd interface{}) (*domain.Booking, error) {
	query := `
		UPDATE bookings
		SET end_date=$2, updated_at=$3
		WHERE id=$1 AND status IN ('pending','confirmed','active')
		RETURNING id, user_id, vehicle_id, status, start_date, end_date,
		          total_price, currency,
		          COALESCE(pickup_location_id::text, ''),
		          COALESCE(dropoff_location_id::text, ''),
		          COALESCE(notes, ''),
		          COALESCE(cancellation_reason, ''),
		          payment_status,
		          COALESCE(payment_ref, ''),
		          created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query, id, newEnd, time.Now())
	return scanBooking(row)
}

func (r *BookingRepo) ConfirmPayment(ctx context.Context, id, paymentMethod, paymentRef string) (*domain.Booking, error) {
	query := `
		UPDATE bookings
		SET status='confirmed', payment_status='paid', payment_ref=$2, updated_at=$3
		WHERE id=$1 AND status='pending'
		RETURNING id, user_id, vehicle_id, status, start_date, end_date,
		          total_price, currency,
		          COALESCE(pickup_location_id::text, ''),
		          COALESCE(dropoff_location_id::text, ''),
		          COALESCE(notes, ''),
		          COALESCE(cancellation_reason, ''),
		          payment_status,
		          COALESCE(payment_ref, ''),
		          created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query, id, paymentRef, time.Now())
	return scanBooking(row)
}

func (r *BookingRepo) ListByVehicleAndDateRange(ctx context.Context, vehicleID string, start, end interface{}) ([]domain.Booking, error) {
	query := `
		SELECT id, user_id, vehicle_id, status, start_date, end_date,
		       total_price, currency,
		       COALESCE(pickup_location_id::text, ''),
		       COALESCE(dropoff_location_id::text, ''),
		       COALESCE(notes, ''),
		       COALESCE(cancellation_reason, ''),
		       payment_status,
		       COALESCE(payment_ref, ''),
		       created_at, updated_at
		FROM bookings
		WHERE vehicle_id = $1
		  AND status NOT IN ('cancelled','completed')
		  AND start_date < $3
		  AND end_date   > $2
		ORDER BY start_date
	`
	rows, err := r.pool.Query(ctx, query, vehicleID, start, end)
	if err != nil {
		return nil, fmt.Errorf("list by vehicle date range: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		b, err := scanBookingRow(rows)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, *b)
	}
	return bookings, nil
}

// ── helpers ─────────────────────────────────────────────────────────────────

func scanBooking(row pgx.Row) (*domain.Booking, error) {
	var b domain.Booking
	err := row.Scan(
		&b.ID, &b.UserID, &b.VehicleID, &b.Status,
		&b.StartDate, &b.EndDate, &b.TotalPrice, &b.Currency,
		&b.PickupLocationID, &b.DropoffLocationID,
		&b.Notes, &b.CancellationReason,
		&b.PaymentStatus, &b.PaymentRef,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("scan booking: %w", err)
	}
	return &b, nil
}

func scanBookingRow(rows pgx.Rows) (*domain.Booking, error) {
	var b domain.Booking
	err := rows.Scan(
		&b.ID, &b.UserID, &b.VehicleID, &b.Status,
		&b.StartDate, &b.EndDate, &b.TotalPrice, &b.Currency,
		&b.PickupLocationID, &b.DropoffLocationID,
		&b.Notes, &b.CancellationReason,
		&b.PaymentStatus, &b.PaymentRef,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan booking row: %w", err)
	}
	return &b, nil
}
