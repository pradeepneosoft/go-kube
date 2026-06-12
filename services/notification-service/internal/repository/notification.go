package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pradeepneosoft/go-kube/services/notification-service/internal/model"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(ctx context.Context, orderID, userID, message string) (model.Notification, error) {
	const query = `
		INSERT INTO notifications (order_id, user_id, message)
		VALUES ($1, $2, $3)
		RETURNING id, order_id, user_id, message, created_at
	`

	var notification model.Notification
	err := r.pool.QueryRow(ctx, query, orderID, userID, message).Scan(
		&notification.ID,
		&notification.OrderID,
		&notification.UserID,
		&notification.Message,
		&notification.CreatedAt,
	)
	if err != nil {
		return model.Notification{}, fmt.Errorf("insert notification: %w", err)
	}

	return notification, nil
}
