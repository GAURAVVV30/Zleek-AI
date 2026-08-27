package application_test

import (
	"context"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/roadmap/application"
	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
	"github.com/jackc/pgx/v5"
)

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

type mockRoadmapRepo struct {
	paths []domain.Path
}

func (m *mockRoadmapRepo) GetActivePath(ctx context.Context, learnerID string) (*domain.Path, []domain.PathItem, error) {
	return nil, nil, domain.ErrActivePathNotFound
}

func (m *mockRoadmapRepo) DeactivatePaths(ctx context.Context, tx pgx.Tx, learnerID string, goalID string) error {
	return nil
}

func (m *mockRoadmapRepo) CreatePath(ctx context.Context, tx pgx.Tx, path *domain.Path, items []domain.PathItem) error {
	m.paths = append(m.paths, *path)
	return nil
}

type mockGoalsSvc struct{}

func (m *mockGoalsSvc) GetActiveGoal(ctx context.Context, learnerID string) (application.Goal, error) {
	return application.Goal{ID: "g-1", KnowledgeStructureID: "ks-1", GoalText: "Learn Go"}, nil
}

type mockKnowledgeSvc struct {
	shouldFail bool
}

func (m *mockKnowledgeSvc) ValidatePrerequisites(ctx context.Context, structureID string, orderedConceptIDs []string) error {
	if m.shouldFail {
		return domain.ErrInvalidPrerequisite
	}
	return nil
}

type mockCompetencySvc struct{}

func (m *mockCompetencySvc) GetLearnerCompetencies(ctx context.Context, learnerID string) (map[string]string, error) {
	return map[string]string{"c-1": "competent"}, nil
}

type mockResourcesSvc struct{}

func (m *mockResourcesSvc) ValidateResources(ctx context.Context, resourceIDs []string) error {
	return nil
}

type mockAISvc struct{}

func (m *mockAISvc) GenerateRoadmapProposal(ctx context.Context, req application.AIProposalRequest) (application.AIProposalResponse, error) {
	res1 := "res-1"
	return application.AIProposalResponse{
		Items: []application.AIProposedItem{
			{ConceptID: "c-1", ResourceID: nil},
			{ConceptID: "c-2", ResourceID: &res1},
		},
	}, nil
}

func (m *mockAISvc) GetConceptExplanation(ctx context.Context, conceptID string) (string, error) {
	return "because", nil
}

func TestRegenerateRoadmap(t *testing.T) {
	repo := &mockRoadmapRepo{}
	uc := application.NewRegenerateRoadmapUseCase(
		&mockTxManager{},
		repo,
		&mockGoalsSvc{},
		&mockKnowledgeSvc{},
		&mockCompetencySvc{},
		&mockResourcesSvc{},
		&mockAISvc{},
	)

	path, items, err := uc.Execute(context.Background(), "l-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path == nil {
		t.Fatal("expected path")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].State != domain.ItemStateCompetent {
		t.Fatalf("expected competent state for first item, got %v", items[0].State)
	}
}

func TestRegenerateRoadmap_InvalidPrerequisite(t *testing.T) {
	repo := &mockRoadmapRepo{}
	uc := application.NewRegenerateRoadmapUseCase(
		&mockTxManager{},
		repo,
		&mockGoalsSvc{},
		&mockKnowledgeSvc{shouldFail: true},
		&mockCompetencySvc{},
		&mockResourcesSvc{},
		&mockAISvc{},
	)

	_, _, err := uc.Execute(context.Background(), "l-1")
	if err != domain.ErrInvalidPrerequisite {
		t.Fatalf("expected ErrInvalidPrerequisite, got %v", err)
	}
}
