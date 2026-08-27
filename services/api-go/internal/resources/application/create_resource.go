package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type CreateResourceUseCase struct {
	repo              ResourceRepository
	conceptValidation ConceptValidationService
}

func NewCreateResourceUseCase(repo ResourceRepository, conceptValidation ConceptValidationService) *CreateResourceUseCase {
	return &CreateResourceUseCase{
		repo:              repo,
		conceptValidation: conceptValidation,
	}
}

type CreateResourceCommand struct {
	URL          string
	Source       *string
	Author       *string
	ResourceType string
	Difficulty   *string
	ConceptIDs   []string
}

func (uc *CreateResourceUseCase) Execute(ctx context.Context, cmd CreateResourceCommand) (*domain.Resource, error) {
	// Validate concepts exist in Knowledge module
	if len(cmd.ConceptIDs) > 0 {
		if err := uc.conceptValidation.ValidateConcepts(ctx, cmd.ConceptIDs); err != nil {
			return nil, domain.ErrConceptNotFound
		}
	}

	id := uuid.New().String()
	now := time.Now()

	resource := &domain.Resource{
		ID:              id,
		URL:             cmd.URL,
		Source:          cmd.Source,
		Author:          cmd.Author,
		ResourceType:    cmd.ResourceType,
		Difficulty:      cmd.Difficulty,
		Status:          domain.StatusCandidate,
		FreshnessStatus: domain.FreshnessUnverified,
		CreatedAt:       now,
	}

	var concepts []domain.ResourceConcept
	for _, cid := range cmd.ConceptIDs {
		concepts = append(concepts, domain.ResourceConcept{
			ResourceID: id,
			ConceptID:  cid,
		})
	}

	if err := uc.repo.CreateResource(ctx, resource, concepts); err != nil {
		return nil, err
	}

	return resource, nil
}
