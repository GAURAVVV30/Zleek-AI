package application

import (
	"context"
)

type GetConceptExplanationUseCase struct {
	aiSvc AIClientService
}

func NewGetConceptExplanationUseCase(aiSvc AIClientService) *GetConceptExplanationUseCase {
	return &GetConceptExplanationUseCase{
		aiSvc: aiSvc,
	}
}

func (uc *GetConceptExplanationUseCase) Execute(ctx context.Context, conceptID string) (*ConceptExplanation, error) {
	return uc.aiSvc.GetConceptExplanation(ctx, conceptID)
}
