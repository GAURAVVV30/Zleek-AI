package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/competency/application"
	"github.com/hcl-backend/services/api-go/internal/competency/domain"
)

type PostgresCompetencyRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCompetencyRepository(db *pgxpool.Pool) *PostgresCompetencyRepository {
	return &PostgresCompetencyRepository{db: db}
}

func (r *PostgresCompetencyRepository) GetByConceptNodeID(ctx context.Context, learnerID, nodeID string) (*domain.CompetencyRecord, error) {
	query := `
		SELECT c.node_id, c.name, cr.state,
		       COALESCE(ev.score::float8, 0),
		       COALESCE(to_char(ev.created_at, 'YYYY-MM-DD'), ''),
		       COALESCE(ad.type, 'quiz'),
		       cr.updated_at
		FROM platform.competency_records cr
		JOIN platform.concepts c ON c.id = cr.concept_id
		LEFT JOIN platform.evidence_records ev ON ev.id = cr.last_evidence_id
		LEFT JOIN platform.assessment_definitions ad ON ad.id = ev.assessment_definition_id
		WHERE cr.learner_id = $1 AND c.node_id = $2
	`
	var rec domain.CompetencyRecord
	err := r.db.QueryRow(ctx, query, learnerID, nodeID).Scan(
		&rec.ConceptID, &rec.ConceptName, &rec.State, &rec.Score, &rec.LastEvidenceDate, &rec.EvidenceType,
		&rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompetencyNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (r *PostgresCompetencyRepository) ListByLearner(ctx context.Context, learnerID string) ([]domain.CompetencyRecord, error) {
	query := `
		SELECT c.node_id, c.name, cr.state,
		       COALESCE(ev.score::float8, 0),
		       COALESCE(to_char(ev.created_at, 'YYYY-MM-DD'), ''),
		       COALESCE(ad.type, 'quiz'),
		       cr.updated_at
		FROM platform.competency_records cr
		JOIN platform.concepts c ON c.id = cr.concept_id
		LEFT JOIN platform.evidence_records ev ON ev.id = cr.last_evidence_id
		LEFT JOIN platform.assessment_definitions ad ON ad.id = ev.assessment_definition_id
		WHERE cr.learner_id = $1
		ORDER BY cr.updated_at DESC
	`
	rows, err := r.db.Query(ctx, query, learnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CompetencyRecord
	for rows.Next() {
		var rec domain.CompetencyRecord
		if err := rows.Scan(&rec.ConceptID, &rec.ConceptName, &rec.State, &rec.Score, &rec.LastEvidenceDate, &rec.EvidenceType, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if out == nil {
		out = []domain.CompetencyRecord{}
	}
	return out, rows.Err()
}

func (r *PostgresCompetencyRepository) GetHistoryByConceptNodeID(ctx context.Context, learnerID, nodeID string) ([]domain.CompetencyHistory, error) {
	query := `
		SELECT
			row_number() OVER (ORDER BY ch.changed_at DESC) AS attempt,
			to_char(ch.changed_at, 'YYYY-MM-DD HH24:MI'),
			COALESCE(ev.score::float8, 0),
			ev.result,
			COALESCE(ad.type, 'quiz'),
			ch.previous_state, ch.new_state, ch.changed_at
		FROM platform.competency_history ch
		JOIN platform.competency_records cr ON cr.learner_id = ch.learner_id AND cr.concept_id = ch.concept_id
		JOIN platform.concepts c ON c.id = ch.concept_id
		JOIN platform.evidence_records ev ON ev.id = ch.evidence_id
		LEFT JOIN platform.assessment_definitions ad ON ad.id = ev.assessment_definition_id
		WHERE ch.learner_id = $1 AND c.node_id = $2
		ORDER BY ch.changed_at DESC
	`
	rows, err := r.db.Query(ctx, query, learnerID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CompetencyHistory
	for rows.Next() {
		var h domain.CompetencyHistory
		var evidenceType string
		if err := rows.Scan(&h.Attempt, &h.Date, &h.Score, &h.Result, &evidenceType, &h.PreviousState, &h.NewState, &h.ChangedAt); err != nil {
			return nil, err
		}
		h.Details = fmt.Sprintf("%s evaluation; competency %s → %s", evidenceType, h.PreviousState, h.NewState)
		out = append(out, h)
	}
	if out == nil {
		out = []domain.CompetencyHistory{}
	}
	return out, rows.Err()
}

func (r *PostgresCompetencyRepository) UpsertWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, nodeID, state, evidenceID string) (string, error) {
	var previous string
	err := r.db.QueryRow(ctx, `
		SELECT cr.state FROM platform.competency_records cr
		JOIN platform.concepts c ON c.id = cr.concept_id
		WHERE cr.learner_id = $1 AND (c.node_id = $2 OR c.node_id LIKE $2 || '_%')`, learnerID, nodeID).Scan(&previous)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		previous = ""
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO platform.competency_records (learner_id, concept_id, state, last_evidence_id, updated_at)
		VALUES ($1, (SELECT id FROM platform.concepts WHERE node_id = $2 OR node_id LIKE $2 || '_%'), $3, $4, $5)
		ON CONFLICT (learner_id, concept_id) DO UPDATE
		SET state = EXCLUDED.state, last_evidence_id = EXCLUDED.last_evidence_id, updated_at = EXCLUDED.updated_at`,
		learnerID, nodeID, state, evidenceID, time.Now().UTC())
	return previous, err
}

func (r *PostgresCompetencyRepository) AppendHistory(ctx context.Context, tx pgx.Tx, h *application.HistoryRow) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO platform.competency_history (id, learner_id, concept_id, previous_state, new_state, evidence_id, changed_at)
		VALUES ($1, $2, (SELECT id FROM platform.concepts WHERE node_id = $3 OR node_id LIKE $3 || '_%'), $4, $5, $6, $7)`,
		h.HistoryID, h.LearnerID, h.NodeID, h.PreviousState, h.NewState, h.EvidenceID, h.ChangedAt)
	return err
}

func (r *PostgresCompetencyRepository) CreateBaseline(ctx context.Context, learnerID, nodeID, state string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO platform.competency_records (learner_id, concept_id, state, updated_at)
		VALUES ($1, (SELECT id FROM platform.concepts WHERE node_id = $2 OR node_id LIKE $2 || '_%'), $3, $4)
		ON CONFLICT (learner_id, concept_id) DO UPDATE SET state = EXCLUDED.state, updated_at = EXCLUDED.updated_at`,
		learnerID, nodeID, state, time.Now().UTC())
	return err
}
