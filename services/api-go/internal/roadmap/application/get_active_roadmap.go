package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

// GetActiveRoadmapUseCase returns the frontend roadmap view.
type GetActiveRoadmapUseCase struct {
	repo RoadmapRepository
}

func NewGetActiveRoadmapUseCase(repo RoadmapRepository) *GetActiveRoadmapUseCase {
	return &GetActiveRoadmapUseCase{repo: repo}
}

func (uc *GetActiveRoadmapUseCase) Execute(ctx context.Context, learnerID string) (*domain.Roadmap, error) {
	return uc.repo.GetRoadmap(ctx, learnerID)
}
