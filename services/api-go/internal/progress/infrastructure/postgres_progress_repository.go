package infrastructure

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/aiengine"
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
	// Build exhaustive candidate strings to match graph node_id, gold module_id, or concept name
	candidates := aiengine.ResolveAllConceptCandidates(event.ConceptID)

	// 1. Resolve path_item_id, concept_id, sequence_order, and path_id from learner's active path
	var pathItemID, conceptID, pathID string
	var currentOrder int

	err := r.db.QueryRow(ctx, `
		WITH target_path AS (
			SELECT id FROM platform.paths
			WHERE (learner_id::text = $1 OR learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
			  AND status = 'active'
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT pi.id, pi.concept_id, pi.sequence_order, pi.path_id
		FROM platform.path_items pi
		JOIN platform.concepts c ON c.id = pi.concept_id
		WHERE pi.path_id = (SELECT id FROM target_path)
		  AND (
		    c.node_id = ANY($2)
		    OR LOWER(c.node_id) = LOWER($3)
		    OR c.id::text = $3
		    OR LOWER(c.name) = LOWER($3)
		    OR EXISTS (
		         SELECT 1 FROM unnest($2::text[]) cand
		         WHERE length(cand) >= 4 AND (
		           LOWER(c.node_id) = LOWER(cand)
		           OR LOWER(c.name) = LOWER(cand)
		           OR (length(cand) >= 8 AND (c.node_id ILIKE cand || '%' OR c.name ILIKE cand || '%'))
		         )
		       )
		  )
		ORDER BY pi.sequence_order ASC
		LIMIT 1`, event.LearnerID, candidates, event.ConceptID).Scan(&pathItemID, &conceptID, &currentOrder, &pathID)

	if err != nil || pathItemID == "" {
		// Scoped lookup for sequence order if event.ConceptID is a numeric string (e.g. "1" or "6")
		_ = r.db.QueryRow(ctx, `
			WITH target_path AS (
				SELECT id FROM platform.paths
				WHERE (learner_id::text = $1 OR learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
				  AND status = 'active'
				ORDER BY created_at DESC
				LIMIT 1
			)
			SELECT pi.id, pi.concept_id, pi.sequence_order, pi.path_id
			FROM platform.path_items pi
			WHERE pi.path_id = (SELECT id FROM target_path)
			  AND pi.sequence_order = NULLIF(regexp_replace($2, '\D', '', 'g'), '')::int
			LIMIT 1`, event.LearnerID, event.ConceptID).Scan(&pathItemID, &conceptID, &currentOrder, &pathID)
	}

	if pathItemID == "" {
		return domain.ErrInvalidEvent
	}

	// 2. BACKEND PREREQUISITE VALIDATION for module completion
	if event.EventType == "marked_reviewed" || event.EventType == "completed" {
		if currentOrder > 1 {
			var prevCompleted bool
			_ = r.db.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM platform.path_items pi
					WHERE pi.path_id = $1::uuid
					  AND pi.sequence_order = $2 - 1
					  AND (
					    pi.state = 'competent'
					    OR EXISTS (
					      SELECT 1 FROM platform.engagement_events ee
					      WHERE ee.path_item_id = pi.id
					        AND ee.event_type IN ('marked_reviewed', 'completed')
					    )
					  )
				)`, pathID, currentOrder).Scan(&prevCompleted)

			if !prevCompleted {
				return domain.ErrPrerequisiteNotMet
			}
		}
	}

	// 3. Write actual completion event for authenticated learner
	_, err = r.db.Exec(ctx, `
		INSERT INTO platform.engagement_events (id, learner_id, path_item_id, event_type, occurred_at)
		VALUES (
			$1, 
			COALESCE((SELECT id FROM platform.users WHERE id::text = $2 OR id = $2::uuid LIMIT 1), $2::uuid), 
			$3::uuid, 
			$4, 
			$5
		)`,
		event.ID, event.LearnerID, pathItemID, event.EventType, event.Timestamp)
	if err != nil {
		return err
	}

	// 4. Update path item state, competency record, and unlock next module
	if event.EventType == "marked_reviewed" || event.EventType == "completed" {
		if conceptID != "" {
			_, _ = r.db.Exec(ctx, `
				INSERT INTO platform.competency_records (learner_id, concept_id, state, updated_at)
				VALUES (
					COALESCE((SELECT id FROM platform.users WHERE id::text = $1 OR id = $1::uuid LIMIT 1), $1::uuid), 
					$2::uuid, 
					'competent', 
					$3
				)
				ON CONFLICT (learner_id, concept_id) DO UPDATE SET state = 'competent', updated_at = EXCLUDED.updated_at`,
				event.LearnerID, conceptID, time.Now().UTC())
		}

		// Update completed module state to competent
		_, _ = r.db.Exec(ctx, `
			UPDATE platform.path_items SET state = 'competent'
			WHERE id = $1::uuid`, pathItemID)

		// Unlock the next module (currentOrder + 1) to available
		_, _ = r.db.Exec(ctx, `
			UPDATE platform.path_items SET state = 'available'
			WHERE path_id = $1::uuid
			  AND sequence_order = $2 + 1
			  AND state = 'locked'`, pathID, currentOrder)
	}

	return nil
}

// Summary computes the dashboard aggregate for a learner's active structure and active personalized path.
func (r *PostgresProgressRepository) Summary(ctx context.Context, learnerID, structureID string) (*application.SummaryPayload, error) {
	if structureID == "" {
		return nil, domain.ErrNoActivePath
	}
	payload := &application.SummaryPayload{Breakdown: []domain.SummaryRow{}}

	// Query concepts from the learner's latest active generated roadmap path for this knowledge structure
	var totalInPath int
	err := r.db.QueryRow(ctx, `
		WITH target_path AS (
			SELECT id FROM platform.paths
			WHERE (learner_id::text = $1 OR learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
			  AND status = 'active'
			  AND (knowledge_structure_id::text = $2 OR knowledge_structure_id = $2::uuid)
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT COUNT(*)
		FROM platform.path_items pi
		WHERE pi.path_id = (SELECT id FROM target_path)`,
		learnerID, structureID).Scan(&totalInPath)

	if err == nil && totalInPath > 0 {
		payload.TotalConcepts = totalInPath
		rows, err := r.db.Query(ctx, `
			WITH target_path AS (
				SELECT id FROM platform.paths
				WHERE (learner_id::text = $1 OR learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
				  AND status = 'active'
				  AND (knowledge_structure_id::text = $2 OR knowledge_structure_id = $2::uuid)
				ORDER BY created_at DESC
				LIMIT 1
			)
			SELECT c.node_id, c.name,
			       CASE 
			         WHEN pi.state = 'competent' THEN 'competent'
			         WHEN EXISTS (
			           SELECT 1 FROM platform.engagement_events ee 
			           WHERE (ee.learner_id::text = $1 OR ee.learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
			             AND ee.path_item_id = pi.id
			             AND ee.event_type IN ('marked_reviewed', 'completed')
			         ) THEN 'competent'
			         WHEN pi.state IS NOT NULL AND pi.state != '' THEN pi.state
			         ELSE 'not_started'
			       END AS state,
			       COALESCE(ev.score::float8, 0)
			FROM platform.path_items pi
			JOIN platform.concepts c ON c.id = pi.concept_id
			LEFT JOIN platform.competency_records cr ON cr.concept_id = c.id AND (cr.learner_id::text = $1 OR cr.learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
			LEFT JOIN platform.evidence_records ev ON ev.id = cr.last_evidence_id
			WHERE pi.path_id = (SELECT id FROM target_path)
			ORDER BY pi.sequence_order ASC`, learnerID, structureID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		prevCompleted := true
		for rows.Next() {
			var nodeID, name, rawState string
			var score float64
			if err := rows.Scan(&nodeID, &name, &rawState, &score); err != nil {
				return nil, err
			}

			evalState := "locked"
			if prevCompleted && (rawState == "competent") {
				evalState = "competent"
				prevCompleted = true
			} else if prevCompleted {
				evalState = "available"
				prevCompleted = false
			} else {
				evalState = "locked"
				prevCompleted = false
			}

			row := domain.SummaryRow{
				Domain:     name,
				Percentage: percentageFor(evalState, score),
				Status:     statusText(evalState),
			}
			payload.Breakdown = append(payload.Breakdown, row)
			switch evalState {
			case "competent":
				payload.Competent++
			case "weak_evidence":
				payload.Weak++
			case "in_progress", "available":
				payload.InProgress++
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	} else {
		// Fallback for learners who have not generated an active path yet
		if err := r.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM platform.concepts WHERE knowledge_structure_id = $1`,
			structureID).Scan(&payload.TotalConcepts); err != nil {
			return nil, err
		}

		rows, err := r.db.Query(ctx, `
			SELECT c.node_id, c.name,
			       CASE 
			         WHEN EXISTS (
			           SELECT 1 FROM platform.engagement_events ee 
			           JOIN platform.path_items pi ON pi.id = ee.path_item_id
			           WHERE (ee.learner_id::text = $1 OR ee.learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
			             AND pi.concept_id = c.id 
			             AND ee.event_type IN ('marked_reviewed', 'completed')
			         ) THEN 'competent'
			         WHEN cr.state IS NOT NULL AND cr.state != 'competent' THEN cr.state
			         ELSE 'not_started'
			       END AS state,
			       COALESCE(ev.score::float8, 0)
			FROM platform.concepts c
			LEFT JOIN platform.competency_records cr ON cr.concept_id = c.id AND (cr.learner_id::text = $1 OR cr.learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
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
	}

	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM platform.competency_records cr
		JOIN platform.concepts c ON c.id = cr.concept_id
		WHERE (cr.learner_id::text = $1 OR cr.learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
		  AND c.knowledge_structure_id = $2 AND cr.state = 'weak_evidence'`,
		learnerID, structureID).Scan(&payload.Remediations); err != nil {
		return nil, err
	}

	actRows, err := r.db.Query(ctx, `
		SELECT to_char(d.day, 'YYYY-MM-DD') AS act_date,
		       COALESCE(e.cnt, 0) AS count
		FROM generate_series(
		    CURRENT_DATE - INTERVAL '364 days',
		    CURRENT_DATE,
		    INTERVAL '1 day'
		) AS d(day)
		LEFT JOIN (
		    SELECT DATE(occurred_at) AS evt_date, COUNT(*) AS cnt
		    FROM platform.engagement_events
		    WHERE (learner_id::text = $1 OR learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
		    GROUP BY DATE(occurred_at)
		) e ON e.evt_date = DATE(d.day)
		ORDER BY d.day ASC`, learnerID)
	if err == nil {
		defer actRows.Close()
		var activities []domain.ActivityDay
		for actRows.Next() {
			var ad domain.ActivityDay
			if err := actRows.Scan(&ad.Date, &ad.Count); err == nil {
				activities = append(activities, ad)
			}
		}
		payload.ActivityData = activities
	} else {
		payload.ActivityData = []domain.ActivityDay{}
	}

	return payload, nil
}

func (r *PostgresProgressRepository) CompetentConceptNames(ctx context.Context, learnerID, structureID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.name FROM platform.competency_records cr
		JOIN platform.concepts c ON c.id = cr.concept_id
		WHERE (cr.learner_id::text = $1 OR cr.learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
		  AND (c.knowledge_structure_id::text = $2 OR c.knowledge_structure_id = $2::uuid)
		  AND cr.state = 'competent'
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

func (r *PostgresProgressRepository) GetCompletionBadgeStatus(ctx context.Context, learnerID, structureID string) (totalModules int, completedModules int, err error) {
	err = r.db.QueryRow(ctx, `
		WITH active_learner_path AS (
			SELECT id FROM platform.paths
			WHERE (learner_id::text = $1 OR learner_id IN (SELECT id FROM platform.users WHERE id::text = $1))
			  AND status = 'active'
			  AND (knowledge_structure_id::text = $2 OR knowledge_structure_id = $2::uuid)
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT 
			COUNT(pi.id) AS total_count,
			COUNT(CASE WHEN pi.state = 'competent' THEN 1 END) AS completed_count
		FROM platform.path_items pi
		WHERE pi.path_id = (SELECT id FROM active_learner_path)`,
		learnerID, structureID).Scan(&totalModules, &completedModules)

	if err != nil {
		return 0, 0, nil
	}
	return totalModules, completedModules, nil
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
