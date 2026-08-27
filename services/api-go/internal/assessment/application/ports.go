package application

import (
	"context"
	"encoding/json"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type AssessmentRepository interface {
	GetDefinitionByConceptID(ctx context.Context, conceptID string) (*domain.AssessmentDefinition, error)
	GetItemsByDefinitionID(ctx context.Context, definitionID string) ([]domain.AssessmentItem, error)
}

type ConceptService interface {
	ValidateConcept(ctx context.Context, conceptID string) error
}

type AIClient interface {
	Evaluate(ctx context.Context, submission json.RawMessage, rubric json.RawMessage) (*domain.EvaluationResult, error)
}

type EvidenceService interface {
	RecordEvidence(ctx context.Context, evidence *domain.Evidence) error
}
