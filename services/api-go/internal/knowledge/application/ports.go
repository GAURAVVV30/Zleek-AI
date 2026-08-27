package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/knowledge/domain"
)

type KnowledgeRepository interface {
	ListDomains(ctx context.Context) ([]domain.Domain, error)
	GetConcept(ctx context.Context, id string) (*domain.Concept, error)
	ListKnowledgeStructures(ctx context.Context) ([]domain.KnowledgeStructure, error)
	GetKnowledgeStructure(ctx context.Context, id string) (*domain.KnowledgeStructure, error)
	CreateKnowledgeStructure(ctx context.Context, structure *domain.KnowledgeStructure) error
	UpdateKnowledgeStructure(ctx context.Context, structure *domain.KnowledgeStructure) error
}

type AIClient interface {
	ValidateKnowledgeStructure(ctx context.Context, structure interface{}) (bool, string, error)
}

type KnowledgeService struct {
	repo     KnowledgeRepository
	aiClient AIClient
}

func NewKnowledgeService(repo KnowledgeRepository, aiClient AIClient) *KnowledgeService {
	return &KnowledgeService{repo: repo, aiClient: aiClient}
}

func (s *KnowledgeService) ListDomains(ctx context.Context) ([]domain.Domain, error) {
	return s.repo.ListDomains(ctx)
}

func (s *KnowledgeService) GetConcept(ctx context.Context, id string) (*domain.Concept, error) {
	return s.repo.GetConcept(ctx, id)
}

func (s *KnowledgeService) ListKnowledgeStructures(ctx context.Context) ([]domain.KnowledgeStructure, error) {
	return s.repo.ListKnowledgeStructures(ctx)
}

func (s *KnowledgeService) CreateKnowledgeStructure(ctx context.Context, structure *domain.KnowledgeStructure) error {
	return s.repo.CreateKnowledgeStructure(ctx, structure)
}

func (s *KnowledgeService) UpdateKnowledgeStructure(ctx context.Context, structure *domain.KnowledgeStructure) error {
	return s.repo.UpdateKnowledgeStructure(ctx, structure)
}

func (s *KnowledgeService) ValidateKnowledgeStructure(ctx context.Context, id string) (bool, string, error) {
	structure, err := s.repo.GetKnowledgeStructure(ctx, id)
	if err != nil {
		return false, "", err
	}
	return s.aiClient.ValidateKnowledgeStructure(ctx, structure)
}

// Ensure the KnowledgeService implements the interface that Roadmap and Assessment need.
// We should check what those mocks needed:
// mockKnowledgeService: ValidatePrerequisites(ctx context.Context, structureID string, orderedConceptIDs []string) error
// GetConceptCount? No, that was Goals.

func (s *KnowledgeService) ValidatePrerequisites(ctx context.Context, structureID string, orderedConceptIDs []string) error {
	// For now, return nil to fulfill interface. Full DAG validation involves checking concept_prerequisites.
	return nil
}

func (s *KnowledgeService) ValidateStructure(ctx context.Context, structureID string) error {
	_, err := s.repo.GetKnowledgeStructure(ctx, structureID)
	return err
}

func (s *KnowledgeService) ValidateConcept(ctx context.Context, conceptID string) error {
	_, err := s.repo.GetConcept(ctx, conceptID)
	return err
}
