package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
	"github.com/jackc/pgx/v5"
)

type RegenerateRoadmapUseCase struct {
	txManager         database.TxManager
	repo              RoadmapRepository
	goalsSvc          GoalsService
	knowledgeSvc      KnowledgeService
	competencySvc     CompetencyService
	resourcesSvc      ResourcesService
	aiSvc             AIClientService
}

func NewRegenerateRoadmapUseCase(
	txManager database.TxManager,
	repo RoadmapRepository,
	goalsSvc GoalsService,
	knowledgeSvc KnowledgeService,
	competencySvc CompetencyService,
	resourcesSvc ResourcesService,
	aiSvc AIClientService,
) *RegenerateRoadmapUseCase {
	return &RegenerateRoadmapUseCase{
		txManager:     txManager,
		repo:          repo,
		goalsSvc:      goalsSvc,
		knowledgeSvc:  knowledgeSvc,
		competencySvc: competencySvc,
		resourcesSvc:  resourcesSvc,
		aiSvc:         aiSvc,
	}
}

func (uc *RegenerateRoadmapUseCase) Execute(ctx context.Context, learnerID string) (*domain.Path, []domain.PathItem, error) {
	goal, err := uc.goalsSvc.GetActiveGoal(ctx, learnerID)
	if err != nil {
		return nil, nil, domain.ErrNoActiveGoal
	}

	competencies, err := uc.competencySvc.GetLearnerCompetencies(ctx, learnerID)
	if err != nil {
		return nil, nil, err
	}

	aiReq := AIProposalRequest{
		GoalText:        goal.GoalText,
		CompetencyState: competencies,
	}

	proposal, err := uc.aiSvc.GenerateRoadmapProposal(ctx, aiReq)
	if err != nil {
		return nil, nil, domain.ErrAIUnavailable
	}
	if len(proposal.Items) == 0 {
		return nil, nil, domain.ErrInvalidAIProposal
	}

	var orderedConceptIDs []string
	var resourceIDs []string
	conceptSet := make(map[string]bool)

	for _, item := range proposal.Items {
		if conceptSet[item.ConceptID] {
			return nil, nil, domain.ErrInvalidAIProposal // Deduplication check
		}
		conceptSet[item.ConceptID] = true
		orderedConceptIDs = append(orderedConceptIDs, item.ConceptID)
		if item.ResourceID != nil {
			resourceIDs = append(resourceIDs, *item.ResourceID)
		}
	}

	if err := uc.knowledgeSvc.ValidatePrerequisites(ctx, goal.KnowledgeStructureID, orderedConceptIDs); err != nil {
		return nil, nil, domain.ErrInvalidPrerequisite
	}

	if len(resourceIDs) > 0 {
		if err := uc.resourcesSvc.ValidateResources(ctx, resourceIDs); err != nil {
			return nil, nil, domain.ErrUnknownResource
		}
	}

	pathID := uuid.New().String()
	now := time.Now()

	newPath := &domain.Path{
		ID:                   pathID,
		LearnerID:            learnerID,
		GoalID:               goal.ID,
		KnowledgeStructureID: goal.KnowledgeStructureID,
		Status:               domain.PathStatusActive,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	var pathItems []domain.PathItem
	for i, aiItem := range proposal.Items {
		state := domain.ItemStateLocked
		if i == 0 {
			state = domain.ItemStateAvailable
		}
		// If learner already competent, maybe mark as competent, but AI shouldn't propose it usually unless remediation.
		if competencies[aiItem.ConceptID] == "competent" {
			state = domain.ItemStateCompetent
		}

		pathItems = append(pathItems, domain.PathItem{
			ID:            uuid.New().String(),
			PathID:        pathID,
			ConceptID:     aiItem.ConceptID,
			ResourceID:    aiItem.ResourceID,
			SequenceOrder: i + 1,
			State:         state,
			IsRemediation: false,
			InsertedAt:    now,
		})
	}

	err = uc.txManager.Do(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		if err := uc.repo.DeactivatePaths(txCtx, tx, learnerID, goal.ID); err != nil {
			return err
		}
		if err := uc.repo.CreatePath(txCtx, tx, newPath, pathItems); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return newPath, pathItems, nil
}
