package application

import (
	"context"
	"encoding/json"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type AssessmentRepository interface {
	GetDefinitionByConceptID(ctx context.Context, conceptID string) (*domain.AssessmentDefinition, error)
	GetItemsByDefinitionID(ctx context.Context, definitionID string) ([]domain.AssessmentItem, error)
	SaveDefinition(ctx context.Context, def *domain.AssessmentDefinition, items []domain.AssessmentItem) error
}

// ConceptCatalog resolves concept identity for assessment generation.
type ConceptCatalog interface {
	ValidateConcept(ctx context.Context, conceptID string) error
	ConceptName(ctx context.Context, conceptID string) (string, error)
	CoreConceptNames(ctx context.Context, conceptID string) ([]string, error)
	ConceptDomain(ctx context.Context, conceptID string) (string, error)
}

type AIClient interface {
	Evaluate(ctx context.Context, conceptID, domainID string, submission json.RawMessage) (*domain.EvaluationResult, error)
}

// EvidenceService records assessment evidence through the progress pipeline.
type EvidenceService interface {
	RecordEvidence(ctx context.Context, evidence *domain.Evidence) (string, error)
}
