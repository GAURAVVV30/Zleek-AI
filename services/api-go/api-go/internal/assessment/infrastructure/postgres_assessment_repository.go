package infrastructure

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAssessmentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAssessmentRepository(db *pgxpool.Pool) *PostgresAssessmentRepository {
	return &PostgresAssessmentRepository{db: db}
}

func (r *PostgresAssessmentRepository) GetDefinitionByConceptID(ctx context.Context, conceptID string) (*domain.AssessmentDefinition, error) {
	query := `
		SELECT id, concept_id, type, rubric, version, generated_by, created_at
		FROM platform.assessment_definitions
		WHERE concept_id = $1
		ORDER BY version DESC
		LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, conceptID)

	var def domain.AssessmentDefinition
	err := row.Scan(&def.ID, &def.ConceptID, &def.Type, &def.Rubric, &def.Version, &def.GeneratedBy, &def.CreatedAt)
	if err != nil {
		return nil, domain.ErrAssessmentNotFound
	}
	return &def, nil
}

func (r *PostgresAssessmentRepository) GetItemsByDefinitionID(ctx context.Context, definitionID string) ([]domain.AssessmentItem, error) {
	query := `
		SELECT id, assessment_definition_id, prompt, item_type, answer_key, created_at
		FROM platform.assessment_items
		WHERE assessment_definition_id = $1
	`
	rows, err := r.db.Query(ctx, query, definitionID)
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
