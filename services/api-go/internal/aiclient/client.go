package aiclient

import (
	"context"
	"encoding/json"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type Client interface {
	ProposeGoalMapping(ctx context.Context, goalText string) (string, error)
	Evaluate(ctx context.Context, submission json.RawMessage, rubric json.RawMessage) (*domain.EvaluationResult, error)
	ValidateKnowledgeStructure(ctx context.Context, structure interface{}) (bool, string, error)
	RankResources(ctx context.Context, conceptID string) ([]string, error)
	ExplainResourceRelevance(ctx context.Context, conceptID, resourceID string) (string, error)
}

type MockClient struct {
	MockProposeGoalMapping func(ctx context.Context, goalText string) (string, error)
	MockEvaluate           func(ctx context.Context, submission json.RawMessage, rubric json.RawMessage) (*domain.EvaluationResult, error)
}

func (m *MockClient) ProposeGoalMapping(ctx context.Context, goalText string) (string, error) {
	if m.MockProposeGoalMapping != nil {
		return m.MockProposeGoalMapping(ctx, goalText)
	}
	// Default mock behavior
	return "mock-knowledge-structure-id", nil
}

func (m *MockClient) Evaluate(ctx context.Context, submission json.RawMessage, rubric json.RawMessage) (*domain.EvaluationResult, error) {
	if m.MockEvaluate != nil {
		return m.MockEvaluate(ctx, submission, rubric)
	}
	// Default mock behavior
	return &domain.EvaluationResult{
		Score:      85.0,
		Confidence: 0.9,
		Result:     "competent",
	}, nil
}

func (m *MockClient) ValidateKnowledgeStructure(ctx context.Context, structure interface{}) (bool, string, error) {
	return true, "mock-validation-passed", nil
}

func (m *MockClient) RankResources(ctx context.Context, conceptID string) ([]string, error) {
	return []string{"mock-res-1", "mock-res-2"}, nil
}

func (m *MockClient) ExplainResourceRelevance(ctx context.Context, conceptID, resourceID string) (string, error) {
	return "AI generated explanation", nil
}
