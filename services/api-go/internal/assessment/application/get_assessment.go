package application

import (
	"context"
	"errors"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

// ensureDefinition loads the latest assessment for a concept, generating the
// deterministic quiz on first access.
func ensureDefinition(ctx context.Context, repo AssessmentRepository, catalog ConceptCatalog, conceptID string) (*domain.AssessmentDefinition, []domain.AssessmentItem, error) {
	def, err := repo.GetDefinitionByConceptID(ctx, conceptID)
	if err == nil {
		items, err := repo.GetItemsByDefinitionID(ctx, def.ID)
		if err != nil {
			return nil, nil, err
		}
		return def, items, nil
	}
	if !errors.Is(err, domain.ErrAssessmentNotFound) {
		return nil, nil, err
	}

	title, _ := catalog.ConceptName(ctx, conceptID)
	core, _ := catalog.CoreConceptNames(ctx, conceptID)
	genDef, items := generateQuiz(conceptID, title, core)
	genDef.ConceptID = conceptID
	if err := repo.SaveDefinition(ctx, genDef, items); err != nil {
		return nil, nil, err
	}
	return genDef, items, nil
}

// GetAssessmentUseCase returns the client-visible quiz for a concept.
type GetAssessmentUseCase struct {
	repo           AssessmentRepository
	conceptCatalog ConceptCatalog
}

func NewGetAssessmentUseCase(repo AssessmentRepository, conceptCatalog ConceptCatalog) *GetAssessmentUseCase {
	return &GetAssessmentUseCase{repo: repo, conceptCatalog: conceptCatalog}
}

func (uc *GetAssessmentUseCase) Execute(ctx context.Context, conceptID string) (*domain.QuizView, error) {
	if err := uc.conceptCatalog.ValidateConcept(ctx, conceptID); err != nil {
		return nil, domain.ErrConceptNotFound
	}
	title, _ := uc.conceptCatalog.ConceptName(ctx, conceptID)
	def, items, err := ensureDefinition(ctx, uc.repo, uc.conceptCatalog, conceptID)
	if err != nil {
		return nil, err
	}
	_ = def
	return &domain.QuizView{
		ConceptID:    conceptID,
		ConceptTitle: title,
		Questions:    questionsForView(items),
	}, nil
}
