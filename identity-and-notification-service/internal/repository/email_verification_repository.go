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

var ErrVerificationNotFound = errors.New("email verification token not found")

type EmailVerificationRepo struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationRepo(pool *pgxpool.Pool) *EmailVerificationRepo {
	return &EmailVerificationRepo{pool: pool}
}

func (r *EmailVerificationRepo) Create(ctx context.Context, ev *domain.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	if ev.ID == "" {
		ev.ID = uuid.New().String()
	}
	_, err := r.pool.Exec(ctx, query, ev.ID, ev.UserID, ev.Token, ev.ExpiresAt, now)
	if err != nil {
		return fmt.Errorf("create email verification: %w", err)
	}
	ev.CreatedAt = now
	return nil
}

func (r *EmailVerificationRepo) GetByToken(ctx context.Context, token string) (*domain.EmailVerification, error) {
	query := `SELECT id, user_id, token, expires_at, used, created_at FROM email_verifications WHERE token = $1`
	row := r.pool.QueryRow(ctx, query, token)
	return scanVerificationRow(row)
}

func (r *EmailVerificationRepo) MarkUsed(ctx context.Context, id string) error {
	query := `UPDATE email_verifications SET used = TRUE WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark verification used: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrVerificationNotFound
	}
	return nil
}

func (r *EmailVerificationRepo) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM email_verifications WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete email verifications: %w", err)
	}
	return nil
}

func scanVerificationRow(row pgx.Row) (*domain.EmailVerification, error) {
	var ev domain.EmailVerification
	err := row.Scan(&ev.ID, &ev.UserID, &ev.Token, &ev.ExpiresAt, &ev.Used, &ev.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVerificationNotFound
		}
		return nil, fmt.Errorf("scan email verification: %w", err)
	}
	return &ev, nil
}
