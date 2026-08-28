package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/assessment/application"
	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type mockAssessmentRepo struct {
	def   *domain.AssessmentDefinition
	items []domain.AssessmentItem
}

func (m *mockAssessmentRepo) GetDefinitionByConceptID(ctx context.Context, conceptID string) (*domain.AssessmentDefinition, error) {
	if conceptID == "invalid" {
		return nil, domain.ErrAssessmentNotFound
	}
	if m.def != nil {
		return m.def, nil
	}
	return &domain.AssessmentDefinition{
		ID:        "def-1",
		ConceptID: conceptID,
		Type:      domain.TypeQuiz,
	}, nil
}

func (m *mockAssessmentRepo) GetItemsByDefinitionID(ctx context.Context, definitionID string) ([]domain.AssessmentItem, error) {
	if m.items != nil {
		return m.items, nil
	}
	return []domain.AssessmentItem{
		{ID: "item-1", AssessmentDefinitionID: definitionID, AnswerKey: json.RawMessage(`{"answer": "A"}`)},
	}, nil
}

type mockConceptService struct{}

func (m *mockConceptService) ValidateConcept(ctx context.Context, conceptID string) error {
	if conceptID == "invalid-concept" {
		return domain.ErrConceptNotFound
	}
	return nil
}

type mockAIClient struct {
	result *domain.EvaluationResult
	err    error
}

func (m *mockAIClient) Evaluate(ctx context.Context, submission json.RawMessage, rubric json.RawMessage) (*domain.EvaluationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &domain.EvaluationResult{Score: 90, Confidence: 0.95, Result: "competent"}, nil
}

type mockEvidenceService struct {
	evidence *domain.Evidence
}

func (m *mockEvidenceService) RecordEvidence(ctx context.Context, evidence *domain.Evidence) error {
	m.evidence = evidence
	return nil
}

func TestGetAssessmentUseCase(t *testing.T) {
	uc := application.NewGetAssessmentUseCase(&mockAssessmentRepo{}, &mockConceptService{})

	def, err := uc.Execute(context.Background(), "concept-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if def == nil {
		t.Fatal("expected definition")
	}
	if len(def.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(def.Items))
	}
	if def.Items[0].AnswerKey != nil {
		t.Errorf("expected answer key to be stripped, got %v", string(def.Items[0].AnswerKey))
	}
}

func TestSubmitAssessmentUseCase_Success(t *testing.T) {
	repo := &mockAssessmentRepo{}
	concept := &mockConceptService{}
	ai := &mockAIClient{}
	evidence := &mockEvidenceService{}

	uc := application.NewSubmitAssessmentUseCase(repo, concept, ai, evidence)

	res, err := uc.Execute(context.Background(), "learner-1", "concept-1", json.RawMessage(`{"q1": "A"}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Result != "competent" {
		t.Errorf("expected competent, got %s", res.Result)
	}
	if evidence.evidence == nil {
		t.Fatal("expected evidence to be recorded")
	}
	if evidence.evidence.Score != 90 {
		t.Errorf("expected score 90, got %f", evidence.evidence.Score)
	}
}

func TestSubmitAssessmentUseCase_InvalidAIResult(t *testing.T) {
	repo := &mockAssessmentRepo{}
	concept := &mockConceptService{}
	ai := &mockAIClient{
		result: &domain.EvaluationResult{Score: 90, Confidence: 0.95, Result: "hallucinated_state"},
	}
	evidence := &mockEvidenceService{}

	uc := application.NewSubmitAssessmentUseCase(repo, concept, ai, evidence)

	_, err := uc.Execute(context.Background(), "learner-1", "concept-1", json.RawMessage(`{"q1": "A"}`))
	if err != domain.ErrInvalidAIResult {
		t.Fatalf("expected ErrInvalidAIResult, got %v", err)
	}
	if evidence.evidence != nil {
		t.Fatal("expected evidence to NOT be recorded for invalid AI result")
	}
}
