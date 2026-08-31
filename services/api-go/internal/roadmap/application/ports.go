package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

type RoadmapRepository interface {
	GetRoadmap(ctx context.Context, learnerID string) (*domain.Roadmap, error)
	DeactivatePaths(ctx context.Context, tx pgx.Tx, learnerID, goalID string) error
	CreatePath(ctx context.Context, tx pgx.Tx, path *domain.Path, items []domain.PathItem) error
	GetDailyTasks(ctx context.Context, learnerID string, start, end time.Time) ([]domain.DailyTaskDay, error)
	SaveDailyTasks(ctx context.Context, learnerID string, tasks []domain.DailyTaskDay) error
	ToggleDailyTask(ctx context.Context, learnerID string, taskID string, completed bool) error
}

type GoalsService interface {
	GetActiveGoal(ctx context.Context, learnerID string) (Goal, error)
}

type Goal struct {
	ID                   string
	KnowledgeStructureID string
	GoalText             string
}

type CompetencyService interface {
	GetLearnerCompetencies(ctx context.Context, learnerID string) (map[string]string, error) // nodeID -> state
}

type AIClientService interface {
	GenerateRoadmapProposal(ctx context.Context, req AIProposalRequest) (AIProposalResponse, error)
	GetConceptExplanation(ctx context.Context, conceptID string) (*ConceptExplanation, error)
}

// ConceptExplanation is the /roadmap/concepts/{id}/why payload.
type ConceptExplanation struct {
	ConceptID        string   `json:"conceptId"`
	ConceptName      string   `json:"conceptName"`
	Reason           string   `json:"reason"`
	PrerequisitesMet []string `json:"prerequisitesMet"`
	UnlocksConcepts  []string `json:"unlocksConcepts"`
}

type AIProposalRequest struct {
	LearnerID       string
	GoalText        string
	CompetencyState map[string]string
}

type AIProposalResponse struct {
	Items []AIProposedItem
}

type AIProposedItem struct {
	ConceptID  string
	ResourceID *string
}
