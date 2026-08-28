package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type ResourceRepository interface {
	GetResource(ctx context.Context, id string) (*domain.Resource, error)
	CreateResource(ctx context.Context, resource *domain.Resource, concepts []domain.ResourceConcept) error
	UpdateResource(ctx context.Context, resource *domain.Resource) error
	ListResources(ctx context.Context) ([]domain.Resource, error)
	GetFeedbackSignals(ctx context.Context, resourceID string) (*domain.ResourceQualitySignal, error)
}

type ConceptValidationService interface {
	ValidateConcepts(ctx context.Context, conceptIDs []string) error
}
