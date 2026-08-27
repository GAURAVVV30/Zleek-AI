package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type AIClient interface {
	RankResources(ctx context.Context, conceptID string) ([]string, error)
	ExplainResourceRelevance(ctx context.Context, conceptID, resourceID string) (string, error)
}

type GetAlternateResourcesUseCase struct {
	repo     ResourceRepository
	aiClient AIClient
}

func NewGetAlternateResourcesUseCase(repo ResourceRepository, aiClient AIClient) *GetAlternateResourcesUseCase {
	return &GetAlternateResourcesUseCase{repo: repo, aiClient: aiClient}
}

func (uc *GetAlternateResourcesUseCase) Execute(ctx context.Context, conceptID string) ([]domain.Resource, error) {
	// 1. Ask AI to rank/suggest alternate resources for this concept
	resourceIDs, err := uc.aiClient.RankResources(ctx, conceptID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch the actual resources from DB
	var resources []domain.Resource
	for _, id := range resourceIDs {
		res, err := uc.repo.GetResource(ctx, id)
		if err == nil {
			resources = append(resources, *res)
		}
	}
	if resources == nil {
		resources = []domain.Resource{}
	}
	return resources, nil
}

type ExplainResourceRelevanceUseCase struct {
	aiClient AIClient
}

func NewExplainResourceRelevanceUseCase(aiClient AIClient) *ExplainResourceRelevanceUseCase {
	return &ExplainResourceRelevanceUseCase{aiClient: aiClient}
}

func (uc *ExplainResourceRelevanceUseCase) Execute(ctx context.Context, conceptID, resourceID string) (string, error) {
	return uc.aiClient.ExplainResourceRelevance(ctx, conceptID, resourceID)
}
