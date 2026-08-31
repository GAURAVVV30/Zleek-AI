package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hcl-backend/services/api-go/internal/feedback/domain"
)

type FeedbackRepository interface {
	RecordFeedback(ctx context.Context, record *domain.FeedbackRecord) error
}

type RecordResourceFeedbackUseCase struct {
	repo FeedbackRepository
}

func NewRecordResourceFeedbackUseCase(repo FeedbackRepository) *RecordResourceFeedbackUseCase {
	return &RecordResourceFeedbackUseCase{repo: repo}
}

func (uc *RecordResourceFeedbackUseCase) Execute(ctx context.Context, learnerID, resourceID string, rating float64, comment string) (*domain.FeedbackRecord, error) {
	if rating < 1 || rating > 5 {
		return nil, domain.ErrInvalidFeedback
	}

	record := &domain.FeedbackRecord{
		ID:         uuid.New().String(),
		LearnerID:  learnerID,
		TargetType: domain.TargetResource,
		TargetID:   resourceID,
		Rating:     rating,
		Comment:    comment,
		CreatedAt:  time.Now().UTC(),
	}

	if err := uc.repo.RecordFeedback(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}
