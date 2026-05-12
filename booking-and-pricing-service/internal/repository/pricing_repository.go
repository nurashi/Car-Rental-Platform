package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/car-rental-booking/internal/domain"
)

var ErrTierNotFound = errors.New("pricing tier not found")

type PricingRepo struct {
	pool *pgxpool.Pool
}

func NewPricingRepo(pool *pgxpool.Pool) *PricingRepo {
	return &PricingRepo{pool: pool}
}

func (r *PricingRepo) GetTiersByCategory(ctx context.Context, vehicleCategory string) ([]domain.PricingTier, error) {
	query := `
		SELECT id, name, vehicle_category, base_daily_rate, seasonal_multiplier,
		       COALESCE(season_start,''), COALESCE(season_end,''), is_active, created_at, updated_at
		FROM pricing_tiers
		WHERE vehicle_category = $1 AND is_active = TRUE
		ORDER BY base_daily_rate
	`
	rows, err := r.pool.Query(ctx, query, vehicleCategory)
	if err != nil {
		return nil, fmt.Errorf("get tiers by category: %w", err)
	}
	defer rows.Close()
	return scanTiers(rows)
}

func (r *PricingRepo) GetAllActive(ctx context.Context) ([]domain.PricingTier, error) {
	query := `
		SELECT id, name, vehicle_category, base_daily_rate, seasonal_multiplier,
		       COALESCE(season_start,''), COALESCE(season_end,''), is_active, created_at, updated_at
		FROM pricing_tiers
		WHERE is_active = TRUE
		ORDER BY vehicle_category, base_daily_rate
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all active tiers: %w", err)
	}
	defer rows.Close()
	return scanTiers(rows)
}

func (r *PricingRepo) GetByID(ctx context.Context, id string) (*domain.PricingTier, error) {
	query := `
		SELECT id, name, vehicle_category, base_daily_rate, seasonal_multiplier,
		       COALESCE(season_start,''), COALESCE(season_end,''), is_active, created_at, updated_at
		FROM pricing_tiers WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	var t domain.PricingTier
	err := row.Scan(
		&t.ID, &t.Name, &t.VehicleCategory, &t.BaseDailyRate,
		&t.SeasonalMultiplier, &t.SeasonStart, &t.SeasonEnd,
		&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTierNotFound
		}
		return nil, fmt.Errorf("get tier by id: %w", err)
	}
	return &t, nil
}

func scanTiers(rows pgx.Rows) ([]domain.PricingTier, error) {
	var tiers []domain.PricingTier
	for rows.Next() {
		var t domain.PricingTier
		if err := rows.Scan(
			&t.ID, &t.Name, &t.VehicleCategory, &t.BaseDailyRate,
			&t.SeasonalMultiplier, &t.SeasonStart, &t.SeasonEnd,
			&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pricing tier: %w", err)
		}
		tiers = append(tiers, t)
	}
	return tiers, nil
}

// ── Refund repository ────────────────────────────────────────────────────────

type RefundRepo struct {
	pool *pgxpool.Pool
}

func NewRefundRepo(pool *pgxpool.Pool) *RefundRepo {
	return &RefundRepo{pool: pool}
}

func (r *RefundRepo) Create(ctx context.Context, ref *domain.Refund) error {
	ref.CreatedAt = time.Now()
	query := `
		INSERT INTO refunds (id, booking_id, amount, reason, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		ref.ID, ref.BookingID, ref.Amount, ref.Reason, ref.Status, ref.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create refund: %w", err)
	}
	return nil
}

func (r *RefundRepo) GetByBookingID(ctx context.Context, bookingID string) ([]domain.Refund, error) {
	query := `SELECT id, booking_id, amount, reason, status, created_at FROM refunds WHERE booking_id = $1`
	rows, err := r.pool.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("get refunds: %w", err)
	}
	defer rows.Close()

	var refunds []domain.Refund
	for rows.Next() {
		var ref domain.Refund
		if err := rows.Scan(&ref.ID, &ref.BookingID, &ref.Amount, &ref.Reason, &ref.Status, &ref.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan refund: %w", err)
		}
		refunds = append(refunds, ref)
	}
	return refunds, nil
}
