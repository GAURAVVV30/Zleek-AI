package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/competency/domain"
)

type GetCompetencyHistoryUseCase struct {
	repo           CompetencyRepository
	conceptService ConceptService
}

func NewGetCompetencyHistoryUseCase(repo CompetencyRepository, conceptService ConceptService) *GetCompetencyHistoryUseCase {
	return &GetCompetencyHistoryUseCase{
		repo:           repo,
		conceptService: conceptService,
	}
}

func (uc *GetCompetencyHistoryUseCase) Execute(ctx context.Context, learnerID, conceptID string) ([]domain.CompetencyHistory, error) {
	if err := uc.conceptService.ValidateConcept(ctx, conceptID); err != nil {
		return nil, domain.ErrConceptNotFound
	}

	history, err := uc.repo.GetHistoryByLearnerAndConcept(ctx, learnerID, conceptID)
	if err != nil {
		return nil, err
	}

	return history, nil
}
