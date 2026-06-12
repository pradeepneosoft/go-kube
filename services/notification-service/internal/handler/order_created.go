package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pradeepneosoft/go-kube/pkg/events"
	"github.com/pradeepneosoft/go-kube/services/notification-service/internal/repository"
)

type OrderCreatedHandler struct {
	repo *repository.NotificationRepository
}

func NewOrderCreatedHandler(repo *repository.NotificationRepository) *OrderCreatedHandler {
	return &OrderCreatedHandler{repo: repo}
}

func (h *OrderCreatedHandler) Handle(ctx context.Context, key, value []byte) error {
	var event events.OrderCreated
	if err := json.Unmarshal(value, &event); err != nil {
		return fmt.Errorf("decode order event: %w", err)
	}

	message := fmt.Sprintf(
		"Order %s created for user %s: %s x%d",
		event.OrderID,
		event.UserID,
		event.ProductName,
		event.Quantity,
	)

	notification, err := h.repo.Create(ctx, event.OrderID, event.UserID, message)
	if err != nil {
		return fmt.Errorf("store notification: %w", err)
	}

	log.Printf("notification stored id=%s order=%s key=%s", notification.ID, event.OrderID, string(key))
	return nil
}
