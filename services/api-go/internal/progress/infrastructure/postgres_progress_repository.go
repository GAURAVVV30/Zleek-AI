package infrastructure

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/progress/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProgressRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProgressRepository(db *pgxpool.Pool) *PostgresProgressRepository {
	return &PostgresProgressRepository{
		db: db,
	}
}

func (r *PostgresProgressRepository) RecordEvidence(ctx context.Context, tx pgx.Tx, evidence *domain.Evidence) error {
	query := `
		INSERT INTO platform.evidence_records (id, learner_id, concept_id, assessment_definition_id, path_item_id, submission_data, score, confidence, evaluator_type, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := tx.Exec(ctx, query,
		evidence.ID,
		evidence.LearnerID,
		evidence.ConceptID,
		evidence.AssessmentDefinitionID,
		evidence.PathItemID,
		evidence.SubmissionData,
		evidence.Score,
		evidence.Confidence,
		evidence.EvaluatorType,
		evidence.Result,
		evidence.CreatedAt,
	)
	return err
}

func (r *PostgresProgressRepository) RecordEngagement(ctx context.Context, event *domain.EngagementEvent) error {
	query := `
		INSERT INTO platform.engagement_events (id, learner_id, path_item_id, event_type, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, event.ID, event.LearnerID, event.PathItemID, event.EventType, event.Timestamp)
	return err
}

func (r *PostgresProgressRepository) GetCompetentCount(ctx context.Context, learnerID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM platform.competency_records
		WHERE learner_id = $1 AND state = 'competent'
	`
	var count int
	err := r.db.QueryRow(ctx, query, learnerID).Scan(&count)
	return count, err
}

func (r *PostgresProgressRepository) GetInProgressCount(ctx context.Context, learnerID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM platform.competency_records
		WHERE learner_id = $1 AND (state = 'in_progress' OR state = 'weak_evidence')
	`
	var count int
	err := r.db.QueryRow(ctx, query, learnerID).Scan(&count)
	return count, err
}
