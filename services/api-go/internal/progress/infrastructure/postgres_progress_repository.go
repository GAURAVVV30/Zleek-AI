package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/progress/application"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type PostgresProgressRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProgressRepository(db *pgxpool.Pool) *PostgresProgressRepository {
	return &PostgresProgressRepository{db: db}
}

// Do runs fn inside a single transaction.
func (r *PostgresProgressRepository) Do(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresProgressRepository) RecordEvidence(ctx context.Context, tx pgx.Tx, evidence *domain.Evidence) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO platform.evidence_records (id, learner_id, concept_id, assessment_definition_id, path_item_id, submission_data, score, confidence, evaluator_type, result, created_at)
		VALUES ($1, $2, (SELECT id FROM platform.concepts WHERE node_id = $3 OR node_id LIKE $3 || '_%'), $4,
		        (SELECT pi.id FROM platform.path_items pi
		          JOIN platform.paths p ON p.id = pi.path_id
		          WHERE p.learner_id = $2 AND p.status = 'active'
		            AND pi.concept_id = (SELECT id FROM platform.concepts WHERE node_id = $3 OR node_id LIKE $3 || '_%')
		          ORDER BY pi.inserted_at DESC LIMIT 1),
		        $5, $6, $7, $8, $9, $10)`,
		evidence.ID, evidence.LearnerID, evidence.ConceptID, evidence.AssessmentDefinitionID,
		evidence.SubmissionData, evidence.Score, evidence.Confidence, evidence.EvaluatorType, evidence.Result,
		evidence.CreatedAt)
	return err
}

func (r *PostgresProgressRepository) SyncPathItemState(ctx context.Context, tx pgx.Tx, learnerID, conceptNodeID, state string) error {
	_, err := tx.Exec(ctx, `
		UPDATE platform.path_items pi
		SET state = $3
		WHERE pi.concept_id = (SELECT id FROM platform.concepts WHERE node_id = $2 OR node_id LIKE $2 || '_%')
		  AND pi.path_id IN (SELECT id FROM platform.paths WHERE learner_id = $1 AND status = 'active')`,
		learnerID, conceptNodeID, state)
	if err != nil {
		return err
	}
	if state != "competent" {
		return nil
	}
	_, err = tx.Exec(ctx, `
		WITH active_path AS (
			SELECT p.id FROM platform.paths p
			JOIN platform.path_items pi ON pi.path_id = p.id
			JOIN platform.concepts c ON c.id = pi.concept_id
			WHERE p.learner_id = $1 AND p.status = 'active'
			  AND (c.node_id = $2 OR c.node_id LIKE $2 || '_%')
			LIMIT 1
		), done AS (
			SELECT pi.sequence_order
			FROM platform.path_items pi
			JOIN platform.concepts c ON c.id = pi.concept_id
			WHERE pi.path_id = (SELECT id FROM active_path) AND (c.node_id = $2 OR c.node_id LIKE $2 || '_%')
		)
		UPDATE platform.path_items pi
		SET state = 'available'
		WHERE pi.path_id = (SELECT id FROM active_path)
		  AND pi.state = 'locked'
		  AND pi.sequence_order = (
		      SELECT MIN(ni.sequence_order)
		      FROM platform.path_items ni
		      WHERE ni.path_id = (SELECT id FROM active_path)
		        AND ni.sequence_order > COALESCE((SELECT sequence_order FROM done), 0)
		  ) AND EXISTS (SELECT 1 FROM done)`,
		learnerID, conceptNodeID)
	return err
}

func (r *PostgresProgressRepository) RecordEngagement(ctx context.Context, event *domain.EngagementEvent) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO platform.engagement_events (id, learner_id, path_item_id, event_type, occurred_at)
		VALUES ($1, $2, (SELECT pi.id FROM platform.path_items pi
		                 JOIN platform.paths p ON p.id = pi.path_id
		                 WHERE p.learner_id = $2 AND p.status = 'active'
		                   AND pi.concept_id = (SELECT id FROM platform.concepts WHERE node_id = $3 OR node_id LIKE $3 || '_%')
		                 ORDER BY pi.inserted_at DESC LIMIT 1), $4, $5)`,
		event.ID, event.LearnerID, event.ConceptID, event.EventType, event.Timestamp)
	return err
}

// Summary computes the dashboard aggregate for a learner's active structure.
func (r *PostgresProgressRepository) Summary(ctx context.Context, learnerID, structureID string) (*application.SummaryPayload, error) {
	if structureID == "" {
		return nil, domain.ErrNoActivePath
	}
	payload := &application.SummaryPayload{Breakdown: []domain.SummaryRow{}}

	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM platform.concepts WHERE knowledge_structure_id = $1`,
		structureID).Scan(&payload.TotalConcepts); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT c.node_id, c.name, COALESCE(cr.state, 'not_started'),
		       COALESCE(ev.score::float8, 0)
		FROM platform.concepts c
		LEFT JOIN platform.competency_records cr ON cr.concept_id = c.id AND cr.learner_id = $1
		LEFT JOIN platform.evidence_records ev ON ev.id = cr.last_evidence_id
		WHERE c.knowledge_structure_id = $2
		ORDER BY c.created_at, c.node_id`, learnerID, structureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, name, state string
		var score float64
		if err := rows.Scan(&nodeID, &name, &state, &score); err != nil {
			return nil, err
		}
		row := domain.SummaryRow{
			Domain:     name,
			Percentage: percentageFor(state, score),
			Status:     statusText(state),
		}
		payload.Breakdown = append(payload.Breakdown, row)
		switch state {
		case "competent":
			payload.Competent++
		case "weak_evidence":
			payload.Weak++
		case "in_progress":
			payload.InProgress++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM platform.competency_records cr
		JOIN platform.concepts c ON c.id = cr.concept_id
		WHERE cr.learner_id = $1 AND c.knowledge_structure_id = $2 AND cr.state = 'weak_evidence'`,
		learnerID, structureID).Scan(&payload.Remediations); err != nil {
		return nil, err
	}
	return payload, nil
}

func (r *PostgresProgressRepository) CompetentConceptNames(ctx context.Context, learnerID, structureID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.name FROM platform.competency_records cr
		JOIN platform.concepts c ON c.id = cr.concept_id
		WHERE cr.learner_id = $1 AND c.knowledge_structure_id = $2 AND cr.state = 'competent'
		ORDER BY c.created_at, c.node_id`, learnerID, structureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

func percentageFor(state string, score float64) int {
	if score > 0 {
		pct := int(score)
		if pct > 100 {
			pct = 100
		}
		return pct
	}
	switch state {
	case "competent":
		return 100
	case "in_progress":
		return 50
	case "weak_evidence":
		return 35
	default:
		return 0
	}
}

func statusText(state string) string {
	switch state {
	case "competent":
		return "Completed"
	case "in_progress":
		return "In Progress"
	case "weak_evidence":
		return "Needs Review"
	default:
		return "Not Started"
	}
}
