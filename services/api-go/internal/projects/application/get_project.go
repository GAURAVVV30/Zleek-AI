package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/projects/domain"
)

type GetProjectUseCase struct {
	assessmentSvc AssessmentService
}

func NewGetProjectUseCase(assessmentSvc AssessmentService) *GetProjectUseCase {
	return &GetProjectUseCase{
		assessmentSvc: assessmentSvc,
	}
}

func (uc *GetProjectUseCase) Execute(ctx context.Context, conceptID string) (*domain.Project, error) {
	return uc.assessmentSvc.GetProjectDefinition(ctx, conceptID)
}
