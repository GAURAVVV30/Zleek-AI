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
	return "machine_learning", nil
}

type mockKnowledgeService struct{}

func (m *mockKnowledgeService) ResolveStructure(ctx context.Context, ref string) (*application.ResolvedStructure, error) {
	if ref != "machine_learning" {
		return nil, domain.ErrGoalNotFound
	}
	return &application.ResolvedStructure{
		ID:          "ks-ml-1",
		DomainSlug:  "machine_learning",
		DomainName:  "Machine Learning",
		Confidence:  0.94,
		IsPublished: true,
	}, nil
}

func TestCreateGoalUseCase_Success(t *testing.T) {
	repo := &mockGoalRepo{}
	aiClient := &mockAIClient{}
	knowledge := &mockKnowledgeService{}
	uc := application.NewCreateGoalUseCase(repo, aiClient, knowledge)

	res, err := uc.Execute(context.Background(), "learner-1", "I want to become an ML engineer")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res == nil {
		t.Fatal("expected mapping result to be returned")
	}
	if res.DomainID != "machine_learning" {
		t.Errorf("expected domain machine_learning, got %s", res.DomainID)
	}
	if res.KnowledgeStructureID != "ks-ml-1" {
		t.Errorf("expected KS ID 'ks-ml-1', got %s", res.KnowledgeStructureID)
	}
	if repo.createdGoal == nil {
		t.Fatal("expected goal to be persisted")
	}
	if repo.createdGoal.KnowledgeStructureID != "ks-ml-1" {
		t.Errorf("expected stored KS 'ks-ml-1', got %s", repo.createdGoal.KnowledgeStructureID)
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
