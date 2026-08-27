package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/goals/domain"
)

type GoalRepository interface {
	Create(ctx context.Context, goal *domain.Goal) error
	GetActiveByLearnerID(ctx context.Context, learnerID string) (*domain.Goal, error)
	Update(ctx context.Context, goal *domain.Goal) error
}

type AIClient interface {
	ProposeGoalMapping(ctx context.Context, goalText string) (string, error)
}

type KnowledgeService interface {
	ValidateStructure(ctx context.Context, structureID string) error
}
