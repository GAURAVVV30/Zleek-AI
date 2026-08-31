package infrastructure

import (
	"context"
	"errors"

	"github.com/hcl-backend/services/api-go/internal/competency/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCompetencyRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCompetencyRepository(db *pgxpool.Pool) *PostgresCompetencyRepository {
	return &PostgresCompetencyRepository{
		db: db,
	}
}

func (r *PostgresCompetencyRepository) GetByLearnerAndConcept(ctx context.Context, learnerID, conceptID string) (*domain.CompetencyRecord, error) {
	query := `
		SELECT learner_id, concept_id, state, last_evidence_id, updated_at
		FROM platform.competency_records
		WHERE learner_id = $1 AND concept_id = $2
	`
	row := r.db.QueryRow(ctx, query, learnerID, conceptID)

	var record domain.CompetencyRecord
	err := row.Scan(&record.LearnerID, &record.ConceptID, &record.State, &record.LastEvidenceID, &record.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompetencyNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *PostgresCompetencyRepository) GetHistoryByLearnerAndConcept(ctx context.Context, learnerID, conceptID string) ([]domain.CompetencyHistory, error) {
	query := `
		SELECT id, learner_id, concept_id, previous_state, new_state, evidence_id, changed_at
		FROM platform.competency_history
		WHERE learner_id = $1 AND concept_id = $2
		ORDER BY changed_at DESC
	`
	rows, err := r.db.Query(ctx, query, learnerID, conceptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []domain.CompetencyHistory
	for rows.Next() {
		var h domain.CompetencyHistory
		if err := rows.Scan(&h.ID, &h.LearnerID, &h.ConceptID, &h.PreviousState, &h.NewState, &h.EvidenceID, &h.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

func (r *PostgresCompetencyRepository) UpsertCompetency(ctx context.Context, tx pgx.Tx, record *domain.CompetencyRecord) error {
	query := `
		INSERT INTO platform.competency_records (learner_id, concept_id, state, last_evidence_id, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (learner_id, concept_id)
		DO UPDATE SET state = EXCLUDED.state, last_evidence_id = EXCLUDED.last_evidence_id, updated_at = EXCLUDED.updated_at
	`
	_, err := tx.Exec(ctx, query, record.LearnerID, record.ConceptID, record.State, record.LastEvidenceID, record.UpdatedAt)
	return err
}

func (r *PostgresCompetencyRepository) AppendHistory(ctx context.Context, tx pgx.Tx, history *domain.CompetencyHistory) error {
	query := `
		INSERT INTO platform.competency_history (id, learner_id, concept_id, previous_state, new_state, evidence_id, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(ctx, query, history.ID, history.LearnerID, history.ConceptID, history.PreviousState, history.NewState, history.EvidenceID, history.ChangedAt)
	return err
}
