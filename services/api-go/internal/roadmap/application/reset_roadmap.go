package application

import (
	"context"
)

type ResetRoadmapUseCase struct {
	repo RoadmapRepository
}

func NewResetRoadmapUseCase(repo RoadmapRepository) *ResetRoadmapUseCase {
	return &ResetRoadmapUseCase{repo: repo}
}

func (uc *ResetRoadmapUseCase) Execute(ctx context.Context, learnerID string) error {
	return uc.repo.ResetRoadmapProgress(ctx, learnerID)
}
