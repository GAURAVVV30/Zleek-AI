package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type RecordEngagementUseCase struct {
	repo ProgressRepository
}

func NewRecordEngagementUseCase(repo ProgressRepository) *RecordEngagementUseCase {
	return &RecordEngagementUseCase{
		repo: repo,
	}
}

func (uc *RecordEngagementUseCase) Execute(ctx context.Context, learnerID, pathItemID, eventType string) error {
	if learnerID == "" || pathItemID == "" || eventType == "" {
		return domain.ErrInvalidEvent
	}

	event := &domain.EngagementEvent{
		ID:         uuid.New().String(),
		LearnerID:  learnerID,
		PathItemID: pathItemID,
		EventType:  eventType,
		Timestamp:  time.Now(),
	}

	return uc.repo.RecordEngagement(ctx, event)
}
