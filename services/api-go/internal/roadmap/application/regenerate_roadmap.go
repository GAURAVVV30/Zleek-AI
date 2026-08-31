package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

type RegenerateRoadmapUseCase struct {
	txManager     database.TxManager
	repo          RoadmapRepository
	goalsSvc      GoalsService
	competencySvc CompetencyService
	aiSvc         AIClientService
}

func NewRegenerateRoadmapUseCase(
	txManager database.TxManager,
	repo RoadmapRepository,
	goalsSvc GoalsService,
	competencySvc CompetencyService,
	aiSvc AIClientService,
) *RegenerateRoadmapUseCase {
	return &RegenerateRoadmapUseCase{
		txManager:     txManager,
		repo:          repo,
		goalsSvc:      goalsSvc,
		competencySvc: competencySvc,
		aiSvc:         aiSvc,
	}
}

func (uc *RegenerateRoadmapUseCase) Execute(ctx context.Context, learnerID string) (*domain.Roadmap, error) {
	goal, err := uc.goalsSvc.GetActiveGoal(ctx, learnerID)
	if err != nil {
		return nil, domain.ErrNoActiveGoal
	}

	competencies, err := uc.competencySvc.GetLearnerCompetencies(ctx, learnerID)
	if err != nil {
		return nil, err
	}

	proposal, err := uc.aiSvc.GenerateRoadmapProposal(ctx, AIProposalRequest{
		LearnerID:       learnerID,
		GoalText:        goal.GoalText,
		CompetencyState: competencies,
	})
	if err != nil {
		return nil, domain.ErrAIUnavailable
	}
	if len(proposal.Items) == 0 {
		return nil, domain.ErrInvalidAIProposal
	}

	pathID := uuid.New().String()
	now := time.Now().UTC()

	path := &domain.Path{
		ID:                   pathID,
		LearnerID:            learnerID,
		GoalID:               goal.ID,
		KnowledgeStructureID: goal.KnowledgeStructureID,
		Status:               domain.PathStatusActive,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	var items []domain.PathItem
	for i, proposed := range proposal.Items {
		state := domain.ItemStateLocked
		if i == 0 {
			state = domain.ItemStateAvailable
		}
		items = append(items, domain.PathItem{
			ID:            uuid.New().String(),
			PathID:        pathID,
			ConceptID:     proposed.ConceptID,
			ResourceID:    nil,
			SequenceOrder: i + 1,
			State:         state,
			IsRemediation: false,
			InsertedAt:    now,
		})
	}

	if err := uc.txManager.Do(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.repo.DeactivatePaths(ctx, tx, learnerID, goal.ID); err != nil {
			return err
		}
		return uc.repo.CreatePath(ctx, tx, path, items)
	}); err != nil {
		return nil, err
	}

	return uc.repo.GetRoadmap(ctx, learnerID)
}
