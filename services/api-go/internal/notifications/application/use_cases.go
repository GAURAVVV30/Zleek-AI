package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/notifications/domain"
)

type GetNotificationsUseCase struct {
	repo NotificationRepository
}

func NewGetNotificationsUseCase(repo NotificationRepository) *GetNotificationsUseCase {
	return &GetNotificationsUseCase{repo: repo}
}

func (uc *GetNotificationsUseCase) Execute(ctx context.Context, learnerID string) ([]domain.Notification, error) {
	return uc.repo.ListForUser(ctx, learnerID)
}

type MarkNotificationReadUseCase struct {
	repo NotificationRepository
}

func NewMarkNotificationReadUseCase(repo NotificationRepository) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{repo: repo}
}

func (uc *MarkNotificationReadUseCase) Execute(ctx context.Context, learnerID, notificationID string) error {
	notif, err := uc.repo.GetByID(ctx, notificationID)
	if err != nil {
		return domain.ErrNotificationNotFound
	}

	if notif.LearnerID != learnerID {
		return domain.ErrUnauthorized
	}

	return uc.repo.MarkRead(ctx, notificationID)
}
