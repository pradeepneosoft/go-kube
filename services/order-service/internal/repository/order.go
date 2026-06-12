package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pradeepneosoft/go-kube/services/order-service/internal/model"
)

var ErrNotFound = errors.New("order not found")

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, userID, productName string, quantity int) (model.Order, error) {
	userID = strings.TrimSpace(userID)
	productName = strings.TrimSpace(productName)

	if _, err := uuid.Parse(userID); err != nil {
		return model.Order{}, fmt.Errorf("invalid user id")
	}
	if productName == "" {
		return model.Order{}, fmt.Errorf("product name is required")
	}
	if quantity <= 0 {
		return model.Order{}, fmt.Errorf("quantity must be greater than zero")
	}

	const query = `
		INSERT INTO orders (user_id, product_name, quantity, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, user_id, product_name, quantity, status, created_at
	`

	var order model.Order
	err := r.pool.QueryRow(ctx, query, userID, productName, quantity).Scan(
		&order.ID,
		&order.UserID,
		&order.ProductName,
		&order.Quantity,
		&order.Status,
		&order.CreatedAt,
	)
	if err != nil {
		return model.Order{}, fmt.Errorf("insert order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (model.Order, error) {
	if _, err := uuid.Parse(id); err != nil {
		return model.Order{}, ErrNotFound
	}

	const query = `
		SELECT id, user_id, product_name, quantity, status, created_at
		FROM orders
		WHERE id = $1
	`

	var order model.Order
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.ProductName,
		&order.Quantity,
		&order.Status,
		&order.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Order{}, ErrNotFound
	}
	if err != nil {
		return model.Order{}, fmt.Errorf("get order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID string) ([]model.Order, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	const query = `
		SELECT id, user_id, product_name, quantity, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.ProductName,
			&order.Quantity,
			&order.Status,
			&order.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}
