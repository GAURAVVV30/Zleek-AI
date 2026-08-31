package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type ListResourcesUseCase struct {
	repo ResourceRepository
}

func NewListResourcesUseCase(repo ResourceRepository) *ListResourcesUseCase {
	return &ListResourcesUseCase{
		repo: repo,
	}
}

func (uc *ListResourcesUseCase) Execute(ctx context.Context) ([]domain.Resource, error) {
	return uc.repo.ListResources(ctx)
}
