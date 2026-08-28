package infrastructure

import (
	"context"
	"errors"

	"github.com/hcl-backend/services/api-go/internal/resources/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresResourceRepository struct {
	db *pgxpool.Pool
}

func NewPostgresResourceRepository(db *pgxpool.Pool) *PostgresResourceRepository {
	return &PostgresResourceRepository{
		db: db,
	}
}

func (r *PostgresResourceRepository) GetResource(ctx context.Context, id string) (*domain.Resource, error) {
	query := `
		SELECT id, url, source, author, resource_type, difficulty, authority_score, provenance_note, status, last_checked_at, freshness_status, curated_by, curated_at, created_at
		FROM platform.resources
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	var resource domain.Resource
	err := row.Scan(
		&resource.ID,
		&resource.URL,
		&resource.Source,
		&resource.Author,
		&resource.ResourceType,
		&resource.Difficulty,
		&resource.AuthorityScore,
		&resource.ProvenanceNote,
		&resource.Status,
		&resource.LastCheckedAt,
		&resource.FreshnessStatus,
		&resource.CuratedBy,
		&resource.CuratedAt,
		&resource.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}
	return &resource, nil
}

func (r *PostgresResourceRepository) CreateResource(ctx context.Context, resource *domain.Resource, concepts []domain.ResourceConcept) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO platform.resources (id, url, source, author, resource_type, difficulty, authority_score, provenance_note, status, last_checked_at, freshness_status, curated_by, curated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err = tx.Exec(ctx, query,
		resource.ID,
		resource.URL,
		resource.Source,
		resource.Author,
		resource.ResourceType,
		resource.Difficulty,
		resource.AuthorityScore,
		resource.ProvenanceNote,
		resource.Status,
		resource.LastCheckedAt,
		resource.FreshnessStatus,
		resource.CuratedBy,
		resource.CuratedAt,
		resource.CreatedAt,
	)
	if err != nil {
		// Could check for unique constraint violation and map to ErrDuplicateResource
		return err
	}

	for _, c := range concepts {
		cQuery := `
			INSERT INTO platform.resource_concepts (resource_id, concept_id, relevance_note)
			VALUES ($1, $2, $3)
		`
		_, err = tx.Exec(ctx, cQuery, c.ResourceID, c.ConceptID, c.RelevanceNote)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresResourceRepository) UpdateResource(ctx context.Context, resource *domain.Resource) error {
	query := `
		UPDATE platform.resources
		SET url = $1, source = $2, author = $3, resource_type = $4, difficulty = $5, authority_score = $6, provenance_note = $7, status = $8, last_checked_at = $9, freshness_status = $10, curated_by = $11, curated_at = $12
		WHERE id = $13
	`
	_, err := r.db.Exec(ctx, query,
		resource.URL,
		resource.Source,
		resource.Author,
		resource.ResourceType,
		resource.Difficulty,
		resource.AuthorityScore,
		resource.ProvenanceNote,
		resource.Status,
		resource.LastCheckedAt,
		resource.FreshnessStatus,
		resource.CuratedBy,
		resource.CuratedAt,
		resource.ID,
	)
	return err
}

func (r *PostgresResourceRepository) ListResources(ctx context.Context) ([]domain.Resource, error) {
	query := `
		SELECT id, url, source, author, resource_type, difficulty, authority_score, provenance_note, status, last_checked_at, freshness_status, curated_by, curated_at, created_at
		FROM platform.resources
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []domain.Resource
	for rows.Next() {
		var res domain.Resource
		if err := rows.Scan(
			&res.ID,
			&res.URL,
			&res.Source,
			&res.Author,
			&res.ResourceType,
			&res.Difficulty,
			&res.AuthorityScore,
			&res.ProvenanceNote,
			&res.Status,
			&res.LastCheckedAt,
			&res.FreshnessStatus,
			&res.CuratedBy,
			&res.CuratedAt,
			&res.CreatedAt,
		); err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, rows.Err()
}

func (r *PostgresResourceRepository) GetFeedbackSignals(ctx context.Context, resourceID string) (*domain.ResourceQualitySignal, error) {
	query := `
		SELECT resource_id, avg_rating, feedback_count, outcome_correlation, updated_at
		FROM platform.resource_quality_signals
		WHERE resource_id = $1
	`
	row := r.db.QueryRow(ctx, query, resourceID)

	var signal domain.ResourceQualitySignal
	err := row.Scan(
		&signal.ResourceID,
		&signal.AvgRating,
		&signal.FeedbackCount,
		&signal.OutcomeCorrelation,
		&signal.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no row exists, return an empty initialized signal
			return &domain.ResourceQualitySignal{
				ResourceID:    resourceID,
				FeedbackCount: 0,
			}, nil
		}
		return nil, err
	}
	return &signal, nil
}
