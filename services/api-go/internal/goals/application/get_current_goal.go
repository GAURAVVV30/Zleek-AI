package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/goals/domain"
)

type GetCurrentGoalUseCase struct {
	repo GoalRepository
}

func NewGetCurrentGoalUseCase(repo GoalRepository) *GetCurrentGoalUseCase {
	return &GetCurrentGoalUseCase{
		repo: repo,
	}
}

func (uc *GetCurrentGoalUseCase) Execute(ctx context.Context, learnerID string) (*domain.Goal, error) {
	return uc.repo.GetActiveByLearnerID(ctx, learnerID)
}
