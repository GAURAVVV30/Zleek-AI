package infrastructure

import (
	"context"
	"errors"

	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRoadmapRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRoadmapRepository(db *pgxpool.Pool) *PostgresRoadmapRepository {
	return &PostgresRoadmapRepository{
		db: db,
	}
}

func (r *PostgresRoadmapRepository) GetActivePath(ctx context.Context, learnerID string) (*domain.Path, []domain.PathItem, error) {
	pathQuery := `
		SELECT id, learner_id, goal_id, knowledge_structure_id, status, created_at, updated_at
		FROM platform.paths
		WHERE learner_id = $1 AND status = 'active'
		LIMIT 1
	`
	var path domain.Path
	err := r.db.QueryRow(ctx, pathQuery, learnerID).Scan(
		&path.ID,
		&path.LearnerID,
		&path.GoalID,
		&path.KnowledgeStructureID,
		&path.Status,
		&path.CreatedAt,
		&path.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrActivePathNotFound
		}
		return nil, nil, err
	}

	itemsQuery := `
		SELECT id, path_id, concept_id, resource_id, sequence_order, state, is_remediation, inserted_at
		FROM platform.path_items
		WHERE path_id = $1
		ORDER BY sequence_order ASC
	`
	rows, err := r.db.Query(ctx, itemsQuery, path.ID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []domain.PathItem
	for rows.Next() {
		var item domain.PathItem
		if err := rows.Scan(
			&item.ID,
			&item.PathID,
			&item.ConceptID,
			&item.ResourceID,
			&item.SequenceOrder,
			&item.State,
			&item.IsRemediation,
			&item.InsertedAt,
		); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}

	return &path, items, nil
}

func (r *PostgresRoadmapRepository) DeactivatePaths(ctx context.Context, tx pgx.Tx, learnerID string, goalID string) error {
	query := `
		UPDATE platform.paths
		SET status = 'completed', updated_at = now()
		WHERE learner_id = $1 AND goal_id = $2 AND status = 'active'
	`
	_, err := tx.Exec(ctx, query, learnerID, goalID)
	return err
}

func (r *PostgresRoadmapRepository) CreatePath(ctx context.Context, tx pgx.Tx, path *domain.Path, items []domain.PathItem) error {
	pathQuery := `
		INSERT INTO platform.paths (id, learner_id, goal_id, knowledge_structure_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(ctx, pathQuery,
		path.ID,
		path.LearnerID,
		path.GoalID,
		path.KnowledgeStructureID,
		path.Status,
		path.CreatedAt,
		path.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for _, item := range items {
		itemQuery := `
			INSERT INTO platform.path_items (id, path_id, concept_id, resource_id, sequence_order, state, is_remediation, inserted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := tx.Exec(ctx, itemQuery,
			item.ID,
			item.PathID,
			item.ConceptID,
			item.ResourceID,
			item.SequenceOrder,
			item.State,
			item.IsRemediation,
			item.InsertedAt,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
