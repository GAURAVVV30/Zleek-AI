package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
	"github.com/jackc/pgx/v5"
)

type RoadmapRepository interface {
	GetActivePath(ctx context.Context, learnerID string) (*domain.Path, []domain.PathItem, error)
	DeactivatePaths(ctx context.Context, tx pgx.Tx, learnerID string, goalID string) error
	CreatePath(ctx context.Context, tx pgx.Tx, path *domain.Path, items []domain.PathItem) error
}

type GoalsService interface {
	GetActiveGoal(ctx context.Context, learnerID string) (Goal, error)
}

type Goal struct {
	ID                   string
	KnowledgeStructureID string
	GoalText             string
}

type KnowledgeService interface {
	ValidatePrerequisites(ctx context.Context, structureID string, orderedConceptIDs []string) error
}

type CompetencyService interface {
	GetLearnerCompetencies(ctx context.Context, learnerID string) (map[string]string, error) // ConceptID -> State
}

type ResourcesService interface {
	ValidateResources(ctx context.Context, resourceIDs []string) error
}

type AIClientService interface {
	GenerateRoadmapProposal(ctx context.Context, req AIProposalRequest) (AIProposalResponse, error)
	GetConceptExplanation(ctx context.Context, conceptID string) (string, error)
}

type AIProposalRequest struct {
	GoalText            string
	CompetencyState     map[string]string
}

type AIProposalResponse struct {
	Items []AIProposedItem
}

type AIProposedItem struct {
	ConceptID  string
	ResourceID *string
}
