package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/feedback/domain"
)

type PostgresFeedbackRepository struct {
	db *pgxpool.Pool
}

func NewPostgresFeedbackRepository(db *pgxpool.Pool) *PostgresFeedbackRepository {
	return &PostgresFeedbackRepository{db: db}
}

func (r *PostgresFeedbackRepository) RecordFeedback(ctx context.Context, record *domain.FeedbackRecord) error {
	query := `
		INSERT INTO platform.feedback_records (id, learner_id, target_type, target_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, record.ID, record.LearnerID, record.TargetType, record.TargetID, record.Rating, record.Comment, record.CreatedAt)
	return err
}
