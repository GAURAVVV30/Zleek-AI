package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/projects/domain"
)

type GetProjectStatusUseCase struct {
	evidenceSvc EvidenceService
}

func NewGetProjectStatusUseCase(evidenceSvc EvidenceService) *GetProjectStatusUseCase {
	return &GetProjectStatusUseCase{
		evidenceSvc: evidenceSvc,
	}
}

func (uc *GetProjectStatusUseCase) Execute(ctx context.Context, learnerID, conceptID string) (*domain.ProjectState, error) {
	return uc.evidenceSvc.GetProjectStatus(ctx, learnerID, conceptID)
}
