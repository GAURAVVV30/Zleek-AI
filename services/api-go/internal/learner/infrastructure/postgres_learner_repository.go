package infrastructure

import (
	"context"
	"encoding/json"
	"errors"

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

func (r *PostgresLearnerProfileRepository) GetProfile(ctx context.Context, learnerID string) (*domain.LearnerProfile, error) {
	query := `
		SELECT learner_id, preferences
		FROM platform.learner_profiles WHERE learner_id = $1
	`
	var profile domain.LearnerProfile
	var prefBytes []byte
	err := r.db.QueryRow(ctx, query, learnerID).Scan(&profile.LearnerID, &prefBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}

	if len(prefBytes) > 0 {
		if err := json.Unmarshal(prefBytes, &profile.Preferences); err != nil {
			return nil, err
		}
	} else {
		profile.Preferences = make(map[string]interface{})
	}

	return &profile, nil
}

func (r *PostgresLearnerProfileRepository) UpsertProfile(ctx context.Context, profile *domain.LearnerProfile) error {
	prefBytes, err := json.Marshal(profile.Preferences)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO platform.learner_profiles (learner_id, preferences)
		VALUES ($1, $2)
		ON CONFLICT (learner_id) DO UPDATE SET preferences = $2, updated_at = now()
	`
	_, err = r.db.Exec(ctx, query, profile.LearnerID, prefBytes)
	return err
}
