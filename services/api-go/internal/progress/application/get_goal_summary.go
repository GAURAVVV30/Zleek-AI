package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type GoalsService interface {
	GetCurrentGoalID(ctx context.Context, learnerID string) (string, error)
	GetGoalConceptCount(ctx context.Context, goalID string) (int, error)
}

type GetGoalCompletionSummaryUseCase struct {
	repo         ProgressRepository
	goalsService GoalsService
}

func NewGetGoalCompletionSummaryUseCase(repo ProgressRepository, goalsService GoalsService) *GetGoalCompletionSummaryUseCase {
	return &GetGoalCompletionSummaryUseCase{
		repo:         repo,
		goalsService: goalsService,
	}
}

func (uc *GetGoalCompletionSummaryUseCase) Execute(ctx context.Context, learnerID string) (*domain.GoalCompletionSummary, error) {
	goalID, err := uc.goalsService.GetCurrentGoalID(ctx, learnerID)
	if err != nil {
		return nil, err
	}

	totalConcepts, err := uc.goalsService.GetGoalConceptCount(ctx, goalID)
	if err != nil {
		return nil, err
	}

	// Ideally this aggregates the intersection of the goal's concepts and the learner's competent concepts.
	// We'll use a naive aggregate for now assuming they've completed some number of competent concepts.
	competent, err := uc.repo.GetCompetentCount(ctx, learnerID)
	if err != nil {
		return nil, err
	}

	completed := competent
	if completed > totalConcepts {
		completed = totalConcepts
	}

	return &domain.GoalCompletionSummary{
		GoalID:        goalID,
		TotalConcepts: totalConcepts,
		Completed:     completed,
	}, nil
}
