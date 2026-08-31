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

// Execute maps the learner goal text onto a roadmap.sh domain, resolves the
// published knowledge structure for it, persists the goal, and returns the
// handler-facing mapping result.
func (uc *CreateGoalUseCase) Execute(ctx context.Context, learnerID, goalText string) (*MappingResult, error) {
	mappedDomainID, err := uc.aiClient.ProposeGoalMapping(ctx, goalText)
	if err != nil {
		return nil, domain.ErrAIProposalInvalid
	}

	resolved, err := uc.knowledge.ResolveStructure(ctx, mappedDomainID)
	if err != nil {
		return nil, err
	}

	existing, err := uc.repo.GetActiveByLearnerID(ctx, learnerID)
	if err == nil && existing != nil {
		existing.Status = domain.StatusAbandoned
		_ = uc.repo.Update(ctx, existing)
	}

	newGoal := &domain.Goal{
		ID:                   uuid.New().String(),
		LearnerID:            learnerID,
		GoalText:             goalText,
		KnowledgeStructureID: resolved.ID,
		Status:               domain.StatusActive,
		CreatedAt:            time.Now().UTC(),
	}
	if err := uc.repo.Create(ctx, newGoal); err != nil {
		return nil, err
	}

	return &MappingResult{
		GoalID:               newGoal.ID,
		DomainID:             resolved.DomainSlug,
		DomainName:           resolved.DomainName,
		Confidence:           resolved.Confidence,
		IsSupported:          true,
		GoalText:             goalText,
		KnowledgeStructureID: resolved.ID,
		Status:               string(newGoal.Status),
		CreatedAt:            newGoal.CreatedAt,
	}, nil
}

type MappingResult struct {
	GoalID               string    `json:"goalId"`
	GoalText             string    `json:"goalText"`
	DomainID             string    `json:"domainId"`
	DomainName           string    `json:"domainName"`
	KnowledgeStructureID string    `json:"knowledgeStructureId"`
	Confidence           float64   `json:"confidence"`
	IsSupported          bool      `json:"isSupported"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"createdAt"`
}

func (uc *CreateGoalUseCase) ResolveStructure(ctx context.Context, ref string) (*ResolvedStructure, error) {
	return uc.knowledge.ResolveStructure(ctx, ref)
}
