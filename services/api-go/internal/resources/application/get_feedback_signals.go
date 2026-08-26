package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type GetFeedbackSignalsUseCase struct {
	repo ResourceRepository
}

func NewGetFeedbackSignalsUseCase(repo ResourceRepository) *GetFeedbackSignalsUseCase {
	return &GetFeedbackSignalsUseCase{
		repo: repo,
	}
}

func (uc *GetFeedbackSignalsUseCase) Execute(ctx context.Context, resourceID string) (*domain.ResourceQualitySignal, error) {
	// Optionally check if resource exists first, or let repo handle it
	return uc.repo.GetFeedbackSignals(ctx, resourceID)
}
