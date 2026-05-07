package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/car-rental-identity/internal/domain"
)

var ErrLicenseNotFound = errors.New("driver license not found")

type LicenseRepo struct {
	pool *pgxpool.Pool
}

func NewLicenseRepo(pool *pgxpool.Pool) *LicenseRepo {
	return &LicenseRepo{pool: pool}
}

func (r *LicenseRepo) Create(ctx context.Context, license *domain.DriverLicense) error {
	query := `
		INSERT INTO driver_licenses (id, user_id, license_number, issuing_country, issue_date, expiry_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	now := time.Now()
	if license.ID == "" {
		license.ID = uuid.New().String()
	}
	_, err := r.pool.Exec(ctx, query,
		license.ID, license.UserID, license.LicenseNumber,
		license.IssuingCountry, license.IssueDate, license.ExpiryDate,
		string(license.Status), now, now,
	)
	if err != nil {
		return fmt.Errorf("create driver license: %w", err)
	}
	license.CreatedAt = now
	license.UpdatedAt = now
	return nil
}

func (r *LicenseRepo) GetByUserID(ctx context.Context, userID string) (*domain.DriverLicense, error) {
	query := `SELECT id, user_id, license_number, issuing_country, issue_date, expiry_date, front_image_url, back_image_url, status, created_at, updated_at FROM driver_licenses WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	row := r.pool.QueryRow(ctx, query, userID)
	return scanLicenseRow(row)
}

func (r *LicenseRepo) Update(ctx context.Context, license *domain.DriverLicense) error {
	query := `
		UPDATE driver_licenses SET status = $2, updated_at = $3 WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, license.ID, string(license.Status), time.Now())
	if err != nil {
		return fmt.Errorf("update driver license: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrLicenseNotFound
	}
	return nil
}

func scanLicenseRow(row pgx.Row) (*domain.DriverLicense, error) {
	var l domain.DriverLicense
	err := row.Scan(
		&l.ID, &l.UserID, &l.LicenseNumber, &l.IssuingCountry,
		&l.IssueDate, &l.ExpiryDate, &l.FrontImageURL, &l.BackImageURL,
		&l.Status, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, fmt.Errorf("scan driver license: %w", err)
	}
	return &l, nil
}
