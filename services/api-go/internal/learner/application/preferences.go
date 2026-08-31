package application

import (
	"context"
	"strings"
	"time"

	"github.com/hcl-backend/services/api-go/internal/learner/domain"
)

const (
	defaultAvatarMale   = "memo-34"
	defaultAvatarFemale = "memo-7"
	avatarBase          = "https://raw.githubusercontent.com/alohe/memojis/main/png"
)

type LearnerProfileRepository interface {
	GetProfile(ctx context.Context, userID string) (*domain.LearnerProfile, error)
	UpsertProfile(ctx context.Context, profile *domain.LearnerProfile) error
}

type PreferencesInput struct {
	TimeAvailability string
	FormatPreference []string
	ExperienceLevel  string
	Gender           string
	AvatarURL        string
	Role             string
}

type UpdatePreferencesUseCase struct {
	repo LearnerProfileRepository
}

func NewUpdatePreferencesUseCase(repo LearnerProfileRepository) *UpdatePreferencesUseCase {
	return &UpdatePreferencesUseCase{repo: repo}
}

func (uc *UpdatePreferencesUseCase) Execute(ctx context.Context, userID string, in PreferencesInput) (*domain.LearnerProfile, error) {
	profile := &domain.LearnerProfile{
		UserID:           userID,
		TimeAvailability: in.TimeAvailability,
		FormatPreference: strings.Join(in.FormatPreference, ","),
		PriorExperience:  in.ExperienceLevel,
		Gender:           in.Gender,
		AvatarURL:        in.AvatarURL,
		Role:             in.Role,
	}

	if existing, err := uc.repo.GetProfile(ctx, userID); err == nil {
		if profile.Role == "" {
			profile.Role = existing.Role
		}
		if profile.TimeAvailability == "" {
			profile.TimeAvailability = existing.TimeAvailability
		}
		if profile.PriorExperience == "" {
			profile.PriorExperience = existing.PriorExperience
		}
		if profile.Gender == "" {
			profile.Gender = existing.Gender
		}
		if profile.AvatarURL == "" {
			profile.AvatarURL = existing.AvatarURL
		}
		if strings.TrimSpace(profile.FormatPreference) == "" {
			profile.FormatPreference = existing.FormatPreference
		}
	}

	if strings.TrimSpace(profile.TimeAvailability) == "" {
		profile.TimeAvailability = domain.Availability5To10
	}
	if strings.TrimSpace(profile.PriorExperience) == "" {
		profile.PriorExperience = domain.ExperienceIntermediate
	}
	if strings.TrimSpace(profile.AvatarURL) == "" {
		avatar := defaultAvatarMale
		if profile.Gender == "female" {
			avatar = defaultAvatarFemale
		}
		profile.AvatarURL = avatarBase + "/" + avatar + ".png"
	}

	if err := uc.repo.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// SettingsRepository updates the display profile stored on the users row.
type SettingsRepository interface {
	UpdateSettings(ctx context.Context, userID, fullName, timezone, theme string) error
}

type UpdateSettingsUseCase struct {
	repo SettingsRepository
}

func NewUpdateSettingsUseCase(repo SettingsRepository) *UpdateSettingsUseCase {
	return &UpdateSettingsUseCase{repo: repo}
}

type SettingsProfile struct {
	FullName  string    `json:"fullName"`
	Timezone  string    `json:"timezone"`
	Theme     string    `json:"theme"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (uc *UpdateSettingsUseCase) Execute(ctx context.Context, userID, fullName, timezone, theme string) error {
	return uc.repo.UpdateSettings(ctx, userID, fullName, timezone, theme)
}
