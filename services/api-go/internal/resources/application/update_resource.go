package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type UpdateResourceUseCase struct {
	repo ResourceRepository
}

func NewUpdateResourceUseCase(repo ResourceRepository) *UpdateResourceUseCase {
	return &UpdateResourceUseCase{
		repo: repo,
	}
}

type UpdateResourceCommand struct {
	ID             string
	URL            *string
	Source         *string
	Author         *string
	ResourceType   *string
	Difficulty     *string
	AuthorityScore *float64
	ProvenanceNote *string
	Status         *string
	CuratorID      *string
}

func (uc *UpdateResourceUseCase) Execute(ctx context.Context, cmd UpdateResourceCommand) (*domain.Resource, error) {
	resource, err := uc.repo.GetResource(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if cmd.URL != nil {
		resource.URL = *cmd.URL
	}
	if cmd.Source != nil {
		resource.Source = cmd.Source
	}
	if cmd.Author != nil {
		resource.Author = cmd.Author
	}
	if cmd.ResourceType != nil {
		resource.ResourceType = *cmd.ResourceType
	}
	if cmd.Difficulty != nil {
		resource.Difficulty = cmd.Difficulty
	}
	if cmd.AuthorityScore != nil {
		resource.AuthorityScore = cmd.AuthorityScore
	}
	if cmd.ProvenanceNote != nil {
		resource.ProvenanceNote = cmd.ProvenanceNote
	}

	if cmd.Status != nil {
		switch domain.ResourceStatus(*cmd.Status) {
		case domain.StatusPublished:
			if cmd.CuratorID == nil {
				return nil, domain.ErrInvalidStateTransition
			}
			if err := resource.Publish(*cmd.CuratorID); err != nil {
				return nil, err
			}
		case domain.StatusRetired:
			if err := resource.Retire(); err != nil {
				return nil, err
			}
		case domain.StatusFlagged, domain.StatusCandidate:
			resource.Status = domain.ResourceStatus(*cmd.Status)
		default:
			return nil, domain.ErrInvalidStateTransition
		}
	}

	if err := uc.repo.UpdateResource(ctx, resource); err != nil {
		return nil, err
	}

	return resource, nil
}
