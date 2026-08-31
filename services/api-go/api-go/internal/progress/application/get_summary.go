package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type GetProgressSummaryUseCase struct {
	repo ProgressRepository
}

func NewGetProgressSummaryUseCase(repo ProgressRepository) *GetProgressSummaryUseCase {
	return &GetProgressSummaryUseCase{
		repo: repo,
	}
}

func (uc *GetProgressSummaryUseCase) Execute(ctx context.Context, learnerID string) (*domain.ProgressSummary, error) {
	competent, err := uc.repo.GetCompetentCount(ctx, learnerID)
	if err != nil {
		return nil, err
	}

	inProgress, err := uc.repo.GetInProgressCount(ctx, learnerID)
	if err != nil {
		return nil, err
	}

	// GoalsCompleted would technically come from Goals, but we can return mock/aggregation 0 for now
	return &domain.ProgressSummary{
		LearnerID:       learnerID,
		CompetentCount:  competent,
		InProgressCount: inProgress,
		GoalsCompleted:  0,
	}, nil
}
