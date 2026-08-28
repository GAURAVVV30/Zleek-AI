package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/knowledge/domain"
)

type PostgresKnowledgeRepository struct {
	db *pgxpool.Pool
}

func NewPostgresKnowledgeRepository(db *pgxpool.Pool) *PostgresKnowledgeRepository {
	return &PostgresKnowledgeRepository{db: db}
}

func (r *PostgresKnowledgeRepository) ListDomains(ctx context.Context) ([]domain.Domain, error) {
	query := `SELECT id, name, description, created_at FROM platform.domains ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []domain.Domain
	for rows.Next() {
		var d domain.Domain
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	if domains == nil {
		domains = []domain.Domain{}
	}
	return domains, nil
}

func (r *PostgresKnowledgeRepository) GetConcept(ctx context.Context, id string) (*domain.Concept, error) {
	query := `
		SELECT id, knowledge_structure_id, title, description, learning_objectives, metadata, created_at
		FROM platform.concepts WHERE id = $1
	`
	var c domain.Concept
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.KnowledgeStructureID, &c.Title, &c.Description, &c.LearningObjectives, &c.Metadata, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConceptNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *PostgresKnowledgeRepository) ListKnowledgeStructures(ctx context.Context) ([]domain.KnowledgeStructure, error) {
	query := `SELECT id, domain_id, title, description, version, status, created_at FROM platform.knowledge_structures ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var structures []domain.KnowledgeStructure
	for rows.Next() {
		var s domain.KnowledgeStructure
		if err := rows.Scan(&s.ID, &s.DomainID, &s.Title, &s.Description, &s.Version, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		structures = append(structures, s)
	}
	if structures == nil {
		structures = []domain.KnowledgeStructure{}
	}
	return structures, nil
}

func (r *PostgresKnowledgeRepository) GetKnowledgeStructure(ctx context.Context, id string) (*domain.KnowledgeStructure, error) {
	query := `SELECT id, domain_id, title, description, version, status, created_at FROM platform.knowledge_structures WHERE id = $1`
	var s domain.KnowledgeStructure
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.DomainID, &s.Title, &s.Description, &s.Version, &s.Status, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrKnowledgeStructureNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresKnowledgeRepository) CreateKnowledgeStructure(ctx context.Context, structure *domain.KnowledgeStructure) error {
	query := `
		INSERT INTO platform.knowledge_structures (id, domain_id, title, description, version, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, structure.ID, structure.DomainID, structure.Title, structure.Description, structure.Version, structure.Status, structure.CreatedAt)
	return err
}

func (r *PostgresKnowledgeRepository) UpdateKnowledgeStructure(ctx context.Context, structure *domain.KnowledgeStructure) error {
	query := `
		UPDATE platform.knowledge_structures
		SET title = $1, description = $2, status = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, structure.Title, structure.Description, structure.Status, structure.ID)
	return err
}
