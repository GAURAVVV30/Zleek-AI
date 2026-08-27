package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type GetAssessmentUseCase struct {
	repo           AssessmentRepository
	conceptService ConceptService
}

func NewGetAssessmentUseCase(repo AssessmentRepository, conceptService ConceptService) *GetAssessmentUseCase {
	return &GetAssessmentUseCase{
		repo:           repo,
		conceptService: conceptService,
	}
}

func (uc *GetAssessmentUseCase) Execute(ctx context.Context, conceptID string) (*domain.AssessmentDefinition, error) {
	if err := uc.conceptService.ValidateConcept(ctx, conceptID); err != nil {
		return nil, domain.ErrConceptNotFound
	}

	def, err := uc.repo.GetDefinitionByConceptID(ctx, conceptID)
	if err != nil {
		return nil, domain.ErrAssessmentNotFound
	}

	items, err := uc.repo.GetItemsByDefinitionID(ctx, def.ID)
	if err != nil {
		return nil, err
	}

	// Strip answer keys for the client response
	for i := range items {
		items[i].AnswerKey = nil
	}
	def.Items = items

	return def, nil
}
