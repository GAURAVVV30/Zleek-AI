package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type scanner interface {
	Scan(dest ...any) error
}

type PostgresAssessmentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAssessmentRepository(db *pgxpool.Pool) *PostgresAssessmentRepository {
	return &PostgresAssessmentRepository{db: db}
}

func (r *PostgresAssessmentRepository) GetDefinitionByConceptID(ctx context.Context, conceptID string) (*domain.AssessmentDefinition, error) {
	query := `
		SELECT ad.id, c.node_id, ad.type, ad.rubric, ad.version, ad.generated_by, ad.created_at
		FROM platform.assessment_definitions ad
		JOIN platform.concepts c ON c.id = ad.concept_id
		WHERE c.node_id = $1 OR c.node_id LIKE $1 || '_%'
		ORDER BY ad.version DESC
		LIMIT 1
	`
	return scanDefinition(r.db.QueryRow(ctx, query, conceptID))
}

func scanDefinition(row scanner) (*domain.AssessmentDefinition, error) {
	var def domain.AssessmentDefinition
	err := row.Scan(&def.ID, &def.ConceptID, &def.Type, &def.Rubric, &def.Version, &def.GeneratedBy, &def.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAssessmentNotFound
		}
		return nil, err
	}
	return &def, nil
}

func (r *PostgresAssessmentRepository) GetItemsByDefinitionID(ctx context.Context, definitionID string) ([]domain.AssessmentItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, assessment_definition_id, prompt, item_type, answer_key, created_at
		FROM platform.assessment_items
		WHERE assessment_definition_id = $1
		ORDER BY created_at`, definitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.AssessmentItem
	for rows.Next() {
		var item domain.AssessmentItem
		if err := rows.Scan(&item.ID, &item.AssessmentDefinitionID, &item.Prompt, &item.ItemType, &item.AnswerKey, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresAssessmentRepository) SaveDefinition(ctx context.Context, def *domain.AssessmentDefinition, items []domain.AssessmentItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	id := uuid.NewString()
	if def.ID != "" {
		id = def.ID
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO platform.assessment_definitions (id, concept_id, type, rubric, version, generated_by, created_at)
		VALUES ($1, (SELECT id FROM platform.concepts WHERE node_id = $2 OR node_id LIKE $2 || '_%'), $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING`,
		id, def.ConceptID, def.Type, def.Rubric, def.Version, def.GeneratedBy, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, item := range items {
		itemID := uuid.NewString()
		if item.ID != "" {
			itemID = item.ID
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO platform.assessment_items (id, assessment_definition_id, prompt, item_type, answer_key, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO NOTHING`,
			itemID, id, item.Prompt, item.ItemType, item.AnswerKey, time.Now().UTC()); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
