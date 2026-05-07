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

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, phone, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	now := time.Now()
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Phone,
		string(user.Status), now, now,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, phone, avatar_url, email_verified, status, created_at, updated_at FROM users WHERE id = $1`
	return r.scanUser(ctx, query, id)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, phone, avatar_url, email_verified, status, created_at, updated_at FROM users WHERE email = $1`
	return r.scanUser(ctx, query, email)
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET first_name = $2, last_name = $3, phone = $4, avatar_url = $5, updated_at = $6
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		user.ID, user.FirstName, user.LastName, user.Phone, user.AvatarURL, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) UpdateEmailVerified(ctx context.Context, userID string, verified bool) error {
	query := `UPDATE users SET email_verified = $2, updated_at = $3 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, userID, verified, time.Now())
	if err != nil {
		return fmt.Errorf("update email verified: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	offset := (page - 1) * pageSize
	if page < 1 {
		offset = 0
	}

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	query := `SELECT id, email, password_hash, first_name, last_name, phone, avatar_url, email_verified, status, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, *u)
	}
	return users, total, nil
}

func (r *UserRepo) BlockUser(ctx context.Context, userID, reason string) error {
	query := `UPDATE users SET status = 'blocked', updated_at = $3 WHERE id = $1 AND status != 'blocked'`
	result, err := r.pool.Exec(ctx, query, userID, reason, time.Now())
	if err != nil {
		return fmt.Errorf("block user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) scanUser(ctx context.Context, query string, args ...any) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	return scanRow(row)
}

func scanRow(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Phone, &u.AvatarURL,
		&u.EmailVerified, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}
