package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

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
	query := `SELECT slug, name, description FROM platform.domains ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Domain
	for rows.Next() {
		var d domain.Domain
		if err := rows.Scan(&d.ID, &d.Name, &d.Description); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if out == nil {
		out = []domain.Domain{}
	}
	return out, nil
}

func (r *PostgresKnowledgeRepository) GetDomainBySlug(ctx context.Context, slug string) (*domain.Domain, error) {
	query := `SELECT slug, name, description FROM platform.domains WHERE slug = $1`
	var d domain.Domain
	err := r.db.QueryRow(ctx, query, slug).Scan(&d.ID, &d.Name, &d.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDomainNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *PostgresKnowledgeRepository) GetDomainByStructure(ctx context.Context, structureID string) (string, string, error) {
	query := `
		SELECT d.slug, d.name
		FROM platform.knowledge_structures k
		JOIN platform.domains d ON d.id = k.domain_id
		WHERE k.id::text = $1`
	var slug, name string
	err := r.db.QueryRow(ctx, query, structureID).Scan(&slug, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", domain.ErrKnowledgeStructureNotFound
		}
		return "", "", err
	}
	return slug, name, nil
}

func (r *PostgresKnowledgeRepository) GetConcept(ctx context.Context, id string) (*domain.Concept, error) {
	query := `
		SELECT c.node_id, c.knowledge_structure_id, c.name, c.description, c.created_at
		FROM platform.concepts c
		WHERE c.node_id = $1 OR c.id::text = $1 OR c.node_id LIKE $1 || '_%'
	`
	return scanConcept(r.db.QueryRow(ctx, query, id))
}

func (r *PostgresKnowledgeRepository) ListStructures(ctx context.Context) ([]domain.KnowledgeStructure, error) {
	query := `
		SELECT k.id, k.domain_id, d.slug, k.version, k.status, k.created_by, k.published_at, k.created_at
		FROM platform.knowledge_structures k
		JOIN platform.domains d ON d.id = k.domain_id
		ORDER BY d.name, k.version DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.KnowledgeStructure
	for rows.Next() {
		var s domain.KnowledgeStructure
		if err := rows.Scan(&s.ID, &s.DomainID, &s.DomainName, &s.Version, &s.Status, &s.CreatedBy, &s.PublishedAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []domain.KnowledgeStructure{}
	}
	return out, nil
}

func (r *PostgresKnowledgeRepository) GetStructure(ctx context.Context, id string) (*domain.KnowledgeStructure, error) {
	query := `
		SELECT k.id, k.domain_id, d.slug, k.version, k.status, k.created_by, k.published_at, k.created_at
		FROM platform.knowledge_structures k
		JOIN platform.domains d ON d.id = k.domain_id
		WHERE k.id::text = $1`
	return scanStructure(r.db.QueryRow(ctx, query, id))
}

func (r *PostgresKnowledgeRepository) GetPublishedStructureForDomain(ctx context.Context, slug string) (*domain.KnowledgeStructure, error) {
	query := `
		SELECT k.id, k.domain_id, d.slug, k.version, k.status, k.created_by, k.published_at, k.created_at
		FROM platform.knowledge_structures k
		JOIN platform.domains d ON d.id = k.domain_id
		WHERE d.slug = $1 AND k.status = 'published'
		ORDER BY k.version DESC
		LIMIT 1`
	return scanStructure(r.db.QueryRow(ctx, query, slug))
}

func (r *PostgresKnowledgeRepository) CreateStructure(ctx context.Context, s *domain.KnowledgeStructure) error {
	query := `
		INSERT INTO platform.knowledge_structures (id, domain_id, version, status, created_by, published_at, created_at)
		VALUES ($1, (SELECT id FROM platform.domains WHERE slug = $2), $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, s.ID, s.DomainID, s.Version, s.Status, s.CreatedBy, s.PublishedAt, s.CreatedAt)
	return err
}

func (r *PostgresKnowledgeRepository) UpdateStructure(ctx context.Context, s *domain.KnowledgeStructure, status string, now time.Time) error {
	query := `UPDATE platform.knowledge_structures SET status = $1, published_at = $2 WHERE id::text = $3`
	_, err := r.db.Exec(ctx, query, status, ptrOrNil(status == "published", now.UTC()), s.ID)
	return err
}

func (r *PostgresKnowledgeRepository) CountConcepts(ctx context.Context, structureID string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM platform.concepts WHERE knowledge_structure_id::text = $1`, structureID).Scan(&n)
	return n, err
}

func (r *PostgresKnowledgeRepository) ListConcepts(ctx context.Context, structureID string) ([]domain.Concept, error) {
	query := `
		SELECT c.node_id, c.knowledge_structure_id, c.name, c.description, c.created_at
		FROM platform.concepts c
		WHERE c.knowledge_structure_id::text = $1
		ORDER BY c.name`
	rows, err := r.db.Query(ctx, query, structureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Concept
	for rows.Next() {
		c, err := scanConcept(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	if out == nil {
		out = []domain.Concept{}
	}
	return out, nil
}

func (r *PostgresKnowledgeRepository) ListConceptResources(ctx context.Context, conceptID string) ([]domain.Resource, error) {
	query := `
		SELECT res.id::text, res.url, res.title, COALESCE(res.author,''), res.resource_type,
		       COALESCE(res.difficulty,''), res.authority_score::float8, res.duration_minutes,
		       res.status, res.curated_by::text, res.curated_at, res.created_at
		FROM platform.resources res
		JOIN platform.resource_concepts rc ON rc.resource_id = res.id
		JOIN platform.concepts c ON c.id = rc.concept_id
		WHERE c.node_id = $1
		ORDER BY res.authority_score DESC NULLS LAST`
	rows, err := r.db.Query(ctx, query, conceptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Resource
	for rows.Next() {
		var rRes domain.Resource
		var auth float64
		var dur *int
		var curatedBy *string
		var curatedAt *time.Time
		if err := rows.Scan(&rRes.ID, &rRes.URL, &rRes.Title, &rRes.Author, &rRes.ResourceType,
			&rRes.Difficulty, &auth, &dur, &rRes.Status, &curatedBy, &curatedAt, &rRes.CreatedAt); err != nil {
			return nil, err
		}
		rRes.AuthorityScore = auth
		rRes.DurationMinutes = dur
		rRes.CuratedBy = curatedBy
		rRes.CuratedAt = curatedAt
		out = append(out, rRes)
	}
	if out == nil {
		out = []domain.Resource{}
	}
	return out, nil
}

func (r *PostgresKnowledgeRepository) ListEdges(ctx context.Context, structureID string) ([]domain.Edge, error) {
	query := `
		SELECT cp.concept_id::text, cp.prerequisite_concept_id::text
		FROM platform.concept_prerequisites cp
		JOIN platform.concepts c ON c.id = cp.concept_id
		WHERE c.knowledge_structure_id::text = $1`
	rows, err := r.db.Query(ctx, query, structureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Edge
	for rows.Next() {
		var e domain.Edge
		if err := rows.Scan(&e.ConceptID, &e.PrerequisiteConceptID); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if out == nil {
		out = []domain.Edge{}
	}
	return out, nil
}

func (r *PostgresKnowledgeRepository) GetFormatPrefs(ctx context.Context, userID string) ([]string, error) {
	var raw string
	err := r.db.QueryRow(ctx, `SELECT COALESCE(format_preference,'') FROM platform.learner_profiles WHERE user_id::text = $1`, userID).Scan(&raw)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	parts := strings.Split(raw, ",")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *PostgresKnowledgeRepository) LookupUserName(ctx context.Context, userID string) (string, error) {
	var name string
	query := `SELECT COALESCE(NULLIF(full_name,''), email) FROM platform.users WHERE id::text = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&name)
	return name, err
}

func (r *PostgresKnowledgeRepository) GetConceptState(ctx context.Context, userID, conceptID string) (string, error) {
	var state string
	err := r.db.QueryRow(ctx, `
		SELECT pi.state
		FROM platform.path_items pi
		JOIN platform.paths p ON p.id = pi.path_id
		JOIN platform.concepts c ON c.id = pi.concept_id
		WHERE p.learner_id::text = $1 AND p.status = 'active'
		  AND (c.id::text = $2 OR c.node_id = $2 OR c.node_id LIKE $2 || '_%')`, userID, conceptID).Scan(&state)
	return state, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanConcept(row scanner) (*domain.Concept, error) {
	var c domain.Concept
	err := row.Scan(&c.ID, &c.KnowledgeStructureID, &c.Name, &c.Description, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConceptNotFound
		}
		return nil, err
	}
	return &c, nil
}

func scanStructure(row scanner) (*domain.KnowledgeStructure, error) {
	var s domain.KnowledgeStructure
	err := row.Scan(&s.ID, &s.DomainID, &s.DomainName, &s.Version, &s.Status, &s.CreatedBy, &s.PublishedAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrKnowledgeStructureNotFound
		}
		return nil, err
	}
	return &s, nil
}

func ptrOrNil(b bool, t time.Time) *time.Time {
	if b {
		return &t
	}
	return nil
}
