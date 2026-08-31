package application

import (
	"context"
	"time"

	"github.com/hcl-backend/services/api-go/internal/knowledge/domain"
)

// KnowledgeRepository is the persistence port.
type KnowledgeRepository interface {
	ListDomains(ctx context.Context) ([]domain.Domain, error)
	GetDomainBySlug(ctx context.Context, slug string) (*domain.Domain, error)
	GetDomainByStructure(ctx context.Context, structureID string) (slug, name string, err error)
	GetConcept(ctx context.Context, id string) (*domain.Concept, error)
	ListStructures(ctx context.Context) ([]domain.KnowledgeStructure, error)
	GetStructure(ctx context.Context, id string) (*domain.KnowledgeStructure, error)
	GetPublishedStructureForDomain(ctx context.Context, slug string) (*domain.KnowledgeStructure, error)
	CreateStructure(ctx context.Context, s *domain.KnowledgeStructure) error
	UpdateStructure(ctx context.Context, s *domain.KnowledgeStructure, status string, now time.Time) error
	CountConcepts(ctx context.Context, structureID string) (int, error)
	ListConcepts(ctx context.Context, structureID string) ([]domain.Concept, error)
	ListConceptResources(ctx context.Context, conceptID string) ([]domain.Resource, error)
	ListEdges(ctx context.Context, structureID string) ([]domain.Edge, error)
	GetFormatPrefs(ctx context.Context, userID string) ([]string, error)
	LookupUserName(ctx context.Context, userID string) (string, error)
	GetConceptState(ctx context.Context, userID, conceptID string) (string, error)
}

type RAGService interface {
	GetRAGContext(domainID, nodeID string) []string
}

// ConceptMetaProvider exposes authoritative roadmap.sh graph metadata for a
// concept node (difficulty, category, estimated effort, prerequisites and
// successors).
type ConceptMetaProvider interface {
	Meta(ctx context.Context, conceptID string) (*ConceptMeta, bool)
}

type ConceptMeta struct {
	DomainID       string
	DomainName     string
	ID             string
	Name           string
	Category       string
	Difficulty     string
	EstimatedHours int
	CoreConcepts   []string
	Prereqs        []string
	Successors     []string
}
