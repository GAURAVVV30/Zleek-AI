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

// AIClient maps a free-text goal onto a roadmap.sh domain.
type AIClient interface {
	ProposeGoalMapping(ctx context.Context, goalText string) (domainID string, err error)
}

// ResolvedStructure is the knowledge-structure row matched to a goal. DomainSlug
// is the roadmap.sh domain key (machine_learning, ...).
type ResolvedStructure struct {
	ID          string
	DomainSlug  string
	DomainName  string
	Confidence  float64
	IsPublished bool
}

// KnowledgeService resolves a roadmap domain slug (or structure uuid) into the
// published knowledge structure.
type KnowledgeService interface {
	ResolveStructure(ctx context.Context, ref string) (*ResolvedStructure, error)
}

// GetCurrentGoalUseCase returns the learner's active goal enriched with the
// domain slug/name of its knowledge structure.
type GetCurrentGoalUseCase struct {
	repo       GoalRepository
	structurer StructureNamer
}

type StructureNamer interface {
	StructureName(ctx context.Context, structureID string) (slug, name string, err error)
}

func NewGetCurrentGoalUseCase(repo GoalRepository, structNamer StructureNamer) *GetCurrentGoalUseCase {
	return &GetCurrentGoalUseCase{repo: repo, structurer: structNamer}
}

type CurrentGoalView struct {
	ID         string
	GoalText   string
	Status     string
	DomainID   string
	DomainName string
}

func (uc *GetCurrentGoalUseCase) Execute(ctx context.Context, learnerID string) (*CurrentGoalView, error) {
	goal, err := uc.repo.GetActiveByLearnerID(ctx, learnerID)
	if err != nil {
		return nil, err
	}
	view := &CurrentGoalView{ID: goal.ID, GoalText: goal.GoalText, Status: string(goal.Status)}
	if uc.structurer != nil {
		if slug, name, nerr := uc.structurer.StructureName(ctx, goal.KnowledgeStructureID); nerr == nil {
			view.DomainID = slug
			view.DomainName = name
		}
	}
	return view, nil
}
