package infrastructure

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/goals/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresGoalRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGoalRepository(db *pgxpool.Pool) *PostgresGoalRepository {
	return &PostgresGoalRepository{
		db: db,
	}
}

func (r *PostgresGoalRepository) Create(ctx context.Context, goal *domain.Goal) error {
	query := `
		INSERT INTO platform.goals (id, learner_id, goal_text, knowledge_structure_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, goal.ID, goal.LearnerID, goal.GoalText, goal.KnowledgeStructureID, goal.Status, goal.CreatedAt)
	if err != nil {
		return err
	}

	profileQuery := `
		INSERT INTO platform.learner_profiles (user_id, role, updated_at)
		VALUES ($1, (
			SELECT d.slug
			FROM platform.knowledge_structures k
			JOIN platform.domains d ON d.id = k.domain_id
			WHERE k.id = $2
		), now())
		ON CONFLICT (user_id) DO UPDATE SET
			role = EXCLUDED.role,
			updated_at = now()
	`
	_, _ = r.db.Exec(ctx, profileQuery, goal.LearnerID, goal.KnowledgeStructureID)
	return nil
}

func (r *PostgresGoalRepository) GetActiveByLearnerID(ctx context.Context, learnerID string) (*domain.Goal, error) {
	query := `
		SELECT id, learner_id, goal_text, knowledge_structure_id, status, achieved_at, created_at
		FROM platform.goals
		WHERE learner_id = $1 AND status = 'active'
		LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, learnerID)

	var goal domain.Goal
	err := row.Scan(
		&goal.ID,
		&goal.LearnerID,
		&goal.GoalText,
		&goal.KnowledgeStructureID,
		&goal.Status,
		&goal.AchievedAt,
		&goal.CreatedAt,
	)
	if err != nil {
		return nil, domain.ErrGoalNotFound
	}
	return &goal, nil
}

func (r *PostgresGoalRepository) Update(ctx context.Context, goal *domain.Goal) error {
	query := `
		UPDATE platform.goals
		SET status = $1, achieved_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, goal.Status, goal.AchievedAt, goal.ID)
	return err
}
