package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/learner/domain"
)

type LearnerProfileRepository interface {
	GetProfile(ctx context.Context, learnerID string) (*domain.LearnerProfile, error)
	UpsertProfile(ctx context.Context, profile *domain.LearnerProfile) error
}

type UpdatePreferencesUseCase struct {
	repo LearnerProfileRepository
}

func NewUpdatePreferencesUseCase(repo LearnerProfileRepository) *UpdatePreferencesUseCase {
	return &UpdatePreferencesUseCase{repo: repo}
}

func (uc *UpdatePreferencesUseCase) Execute(ctx context.Context, learnerID string, preferences map[string]interface{}) (*domain.LearnerProfile, error) {
	profile, err := uc.repo.GetProfile(ctx, learnerID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			profile = &domain.LearnerProfile{
				LearnerID:   learnerID,
				Preferences: preferences,
			}
		} else {
			return nil, err
		}
	} else {
		// Merge preferences
		if profile.Preferences == nil {
			profile.Preferences = make(map[string]interface{})
		}
		for k, v := range preferences {
			profile.Preferences[k] = v
		}
	}

	if err := uc.repo.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
