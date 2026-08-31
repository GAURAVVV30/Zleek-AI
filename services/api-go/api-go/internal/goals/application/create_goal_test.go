package application_test

import (
	"context"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/goals/application"
	"github.com/hcl-backend/services/api-go/internal/goals/domain"
)

type mockGoalRepo struct {
	createdGoal *domain.Goal
}

func (m *mockGoalRepo) Create(ctx context.Context, goal *domain.Goal) error {
	m.createdGoal = goal
	return nil
}

func (m *mockGoalRepo) GetActiveByLearnerID(ctx context.Context, learnerID string) (*domain.Goal, error) {
	return nil, domain.ErrGoalNotFound
}

func (m *mockGoalRepo) Update(ctx context.Context, goal *domain.Goal) error {
	return nil
}

type mockAIClient struct{}

func (m *mockAIClient) ProposeGoalMapping(ctx context.Context, goalText string) (string, error) {
	if goalText == "invalid" {
		return "", domain.ErrAIProposalInvalid
	}
	return "mock-ks-id", nil
}

type mockKnowledgeService struct{}

func (m *mockKnowledgeService) ValidateStructure(ctx context.Context, structureID string) error {
	if structureID != "mock-ks-id" {
		return domain.ErrKnowledgeUnpublished
	}
	return nil
}

func TestCreateGoalUseCase_Success(t *testing.T) {
	repo := &mockGoalRepo{}
	aiClient := &mockAIClient{}
	knowledge := &mockKnowledgeService{}
	uc := application.NewCreateGoalUseCase(repo, aiClient, knowledge)

	goal, err := uc.Execute(context.Background(), "learner-1", "Learn Go")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if goal == nil {
		t.Fatal("expected goal to be returned")
	}
	if goal.GoalText != "Learn Go" {
		t.Errorf("expected goal text 'Learn Go', got %s", goal.GoalText)
	}
	if goal.KnowledgeStructureID != "mock-ks-id" {
		t.Errorf("expected KS ID 'mock-ks-id', got %s", goal.KnowledgeStructureID)
	}
	if repo.createdGoal == nil {
		t.Fatal("expected goal to be persisted")
	}
}

func TestCreateGoalUseCase_AIError(t *testing.T) {
	repo := &mockGoalRepo{}
	aiClient := &mockAIClient{}
	knowledge := &mockKnowledgeService{}
	uc := application.NewCreateGoalUseCase(repo, aiClient, knowledge)

	_, err := uc.Execute(context.Background(), "learner-1", "invalid")
	if err != domain.ErrAIProposalInvalid {
		t.Fatalf("expected ErrAIProposalInvalid, got %v", err)
	}
	if repo.createdGoal != nil {
		t.Fatal("expected goal to not be persisted")
	}
}
