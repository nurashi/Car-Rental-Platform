package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/car-rental-identity/internal/domain"
)

type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

func (r *NotificationRepo) Create(ctx context.Context, n *domain.NotificationRecord) error {
	query := `
		INSERT INTO notifications (id, user_id, type, subject, body, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	_, err := r.pool.Exec(ctx, query, n.ID, n.UserID, n.Type, n.Subject, n.Body, time.Now())
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *NotificationRepo) GetByUserID(ctx context.Context, userID string, limit int) ([]domain.NotificationRecord, error) {
	query := `SELECT id, user_id, type, subject, body, created_at FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []domain.NotificationRecord
	for rows.Next() {
		var n domain.NotificationRecord
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Subject, &n.Body, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}
