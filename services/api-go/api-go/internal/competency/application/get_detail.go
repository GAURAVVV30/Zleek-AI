package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/competency/domain"
)

type GetCompetencyDetailUseCase struct {
	repo           CompetencyRepository
	conceptService ConceptService
}

func NewGetCompetencyDetailUseCase(repo CompetencyRepository, conceptService ConceptService) *GetCompetencyDetailUseCase {
	return &GetCompetencyDetailUseCase{
		repo:           repo,
		conceptService: conceptService,
	}
}

func (uc *GetCompetencyDetailUseCase) Execute(ctx context.Context, learnerID, conceptID string) (*domain.CompetencyRecord, error) {
	if err := uc.conceptService.ValidateConcept(ctx, conceptID); err != nil {
		return nil, domain.ErrConceptNotFound
	}

	record, err := uc.repo.GetByLearnerAndConcept(ctx, learnerID, conceptID)
	if err != nil {
		// If no record exists, return a "not_started" state instead of 404
		return &domain.CompetencyRecord{
			LearnerID: learnerID,
			ConceptID: conceptID,
			State:     domain.StateNotStarted,
		}, nil
	}

	return record, nil
}
