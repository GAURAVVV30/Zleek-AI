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
	return err
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
