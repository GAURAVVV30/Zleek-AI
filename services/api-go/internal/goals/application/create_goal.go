package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hcl-backend/services/api-go/internal/goals/domain"
)

type CreateGoalUseCase struct {
	repo      GoalRepository
	aiClient  AIClient
	knowledge KnowledgeService
}

func NewCreateGoalUseCase(repo GoalRepository, aiClient AIClient, knowledge KnowledgeService) *CreateGoalUseCase {
	return &CreateGoalUseCase{
		repo:      repo,
		aiClient:  aiClient,
		knowledge: knowledge,
	}
}

func (uc *CreateGoalUseCase) Execute(ctx context.Context, learnerID, goalText string) (*domain.Goal, error) {
	// 1. Ask FastAPI (AI) to propose a knowledge structure ID based on the text.
	proposedKSID, err := uc.aiClient.ProposeGoalMapping(ctx, goalText)
	if err != nil {
		return nil, domain.ErrAIProposalInvalid
	}

	// 2. Validate the proposed structure exists in Go.
	if err := uc.knowledge.ValidateStructure(ctx, proposedKSID); err != nil {
		return nil, err
	}

	// 3. Mark existing active goal as abandoned (if any).
	existing, err := uc.repo.GetActiveByLearnerID(ctx, learnerID)
	if err == nil && existing != nil {
		existing.Status = domain.StatusAbandoned
		_ = uc.repo.Update(ctx, existing)
	}

	// 4. Create and persist the new goal.
	newGoal := &domain.Goal{
		ID:                   uuid.New().String(),
		LearnerID:            learnerID,
		GoalText:             goalText,
		KnowledgeStructureID: proposedKSID,
		Status:               domain.StatusActive,
		CreatedAt:            time.Now(),
	}

	if err := uc.repo.Create(ctx, newGoal); err != nil {
		return nil, err
	}

	return newGoal, nil
}
