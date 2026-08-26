package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

type GetActiveRoadmapUseCase struct {
	repo RoadmapRepository
}

func NewGetActiveRoadmapUseCase(repo RoadmapRepository) *GetActiveRoadmapUseCase {
	return &GetActiveRoadmapUseCase{
		repo: repo,
	}
}

func (uc *GetActiveRoadmapUseCase) Execute(ctx context.Context, learnerID string) (*domain.Path, []domain.PathItem, error) {
	return uc.repo.GetActivePath(ctx, learnerID)
}
