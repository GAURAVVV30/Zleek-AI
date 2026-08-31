package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/learner/domain"
)

type PostgresLearnerProfileRepository struct {
	db *pgxpool.Pool
}

func NewPostgresLearnerProfileRepository(db *pgxpool.Pool) *PostgresLearnerProfileRepository {
	return &PostgresLearnerProfileRepository{db: db}
}

func (r *PostgresLearnerProfileRepository) GetProfile(ctx context.Context, userID string) (*domain.LearnerProfile, error) {
	query := `
		SELECT user_id, time_availability, format_preference, prior_experience, gender, avatar_url, role
		FROM platform.learner_profiles WHERE user_id = $1
	`
	var profile domain.LearnerProfile
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.UserID, &profile.TimeAvailability, &profile.FormatPreference,
		&profile.PriorExperience, &profile.Gender, &profile.AvatarURL, &profile.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *PostgresLearnerProfileRepository) UpsertProfile(ctx context.Context, profile *domain.LearnerProfile) error {
	query := `
		INSERT INTO platform.learner_profiles
			(user_id, time_availability, format_preference, prior_experience, gender, avatar_url, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			time_availability = EXCLUDED.time_availability,
			format_preference = EXCLUDED.format_preference,
			prior_experience = EXCLUDED.prior_experience,
			gender = EXCLUDED.gender,
			avatar_url = EXCLUDED.avatar_url,
			role = EXCLUDED.role,
			updated_at = $8
	`
	_, err := r.db.Exec(ctx, query, profile.UserID, profile.TimeAvailability, profile.FormatPreference,
		profile.PriorExperience, profile.Gender, profile.AvatarURL, profile.Role, time.Now().UTC())
	return err
}

type PostgresSettingsRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSettingsRepository(db *pgxpool.Pool) *PostgresSettingsRepository {
	return &PostgresSettingsRepository{db: db}
}

func (r *PostgresSettingsRepository) UpdateSettings(ctx context.Context, userID, fullName, timezone, theme string) error {
	query := `
		UPDATE platform.users
		SET full_name = $1, timezone = $2, theme = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, fullName, timezone, theme, time.Now().UTC(), userID)
	return err
}
