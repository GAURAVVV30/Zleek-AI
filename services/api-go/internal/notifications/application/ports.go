package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/notifications/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	ListForUser(ctx context.Context, learnerID string) ([]domain.Notification, error)
	GetByID(ctx context.Context, id string) (*domain.Notification, error)
	MarkRead(ctx context.Context, id string) error
}
