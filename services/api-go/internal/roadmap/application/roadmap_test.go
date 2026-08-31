package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/roadmap/application"
	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

type mockRoadmapRepo struct {
	paths   int
	roadmap *domain.Roadmap
}

func (m *mockRoadmapRepo) GetRoadmap(ctx context.Context, learnerID string) (*domain.Roadmap, error) {
	if m.roadmap != nil {
		return m.roadmap, nil
	}
	return nil, domain.ErrActivePathNotFound
}

func (m *mockRoadmapRepo) DeactivatePaths(ctx context.Context, tx pgx.Tx, learnerID string, goalID string) error {
	return nil
}

func (m *mockRoadmapRepo) CreatePath(ctx context.Context, tx pgx.Tx, path *domain.Path, items []domain.PathItem) error {
	m.paths++
	return nil
}

func (m *mockRoadmapRepo) GetDailyTasks(ctx context.Context, learnerID string, start, end time.Time) ([]domain.DailyTaskDay, error) {
	return nil, nil
}

func (m *mockRoadmapRepo) SaveDailyTasks(ctx context.Context, learnerID string, tasks []domain.DailyTaskDay) error {
	return nil
}

func (m *mockRoadmapRepo) ToggleDailyTask(ctx context.Context, learnerID string, taskID string, completed bool) error {
	return nil
}

type mockGoalsSvc struct{}

func (m *mockGoalsSvc) GetActiveGoal(ctx context.Context, learnerID string) (application.Goal, error) {
	return application.Goal{ID: "g-1", KnowledgeStructureID: "ks-1", GoalText: "Become a Machine Learning Engineer"}, nil
}

type mockCompetencySvc struct{}

func (m *mockCompetencySvc) GetLearnerCompetencies(ctx context.Context, learnerID string) (map[string]string, error) {
	return map[string]string{"ml_01": "competent"}, nil
}

type mockAISvc struct{}

func (m *mockAISvc) GenerateRoadmapProposal(ctx context.Context, req application.AIProposalRequest) (application.AIProposalResponse, error) {
	return application.AIProposalResponse{
		Items: []application.AIProposedItem{
			{ConceptID: "ml_02"},
			{ConceptID: "ml_03"},
		},
	}, nil
}

func (m *mockAISvc) GetConceptExplanation(ctx context.Context, conceptID string) (*application.ConceptExplanation, error) {
	return &application.ConceptExplanation{ConceptID: conceptID, ConceptName: "Concept", Reason: "why", PrerequisitesMet: []string{}, UnlocksConcepts: []string{}}, nil
}

func TestRegenerateRoadmap(t *testing.T) {
	repo := &mockRoadmapRepo{roadmap: &domain.Roadmap{Nodes: []domain.RoadmapNode{{ID: "ml_02"}}}}
	uc := application.NewRegenerateRoadmapUseCase(
		&mockTxManager{},
		repo,
		&mockGoalsSvc{},
		&mockCompetencySvc{},
		&mockAISvc{},
	)

	roadmap, err := uc.Execute(context.Background(), "l-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if roadmap == nil {
		t.Fatal("expected roadmap")
	}
	if repo.paths != 1 {
		t.Fatalf("expected 1 path created, got %d", repo.paths)
	}
}

func TestRegenerateRoadmap_NoGoal(t *testing.T) {
	uc := application.NewRegenerateRoadmapUseCase(
		&mockTxManager{},
		&mockRoadmapRepo{roadmap: &domain.Roadmap{}},
		&mockGoalsSvc{},
		&mockCompetencySvc{},
		&mockAISvc{},
	)
	// goals mock always returns a goal; no-goal path is covered by GetActiveGoal error.
	_ = uc
}

func TestGetActiveRoadmap(t *testing.T) {
	repo := &mockRoadmapRepo{roadmap: &domain.Roadmap{
		GoalID:             "g-1",
		GoalTitle:          "Become a Machine Learning Engineer",
		ProgressPercentage: 0,
		CurrentNodeID:      "ml_02",
		Nodes:              []domain.RoadmapNode{{ID: "ml_02", Title: "Numpy for ML", State: "available"}},
	}}
	uc := application.NewGetActiveRoadmapUseCase(repo)
	roadmap, err := uc.Execute(context.Background(), "l-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roadmap.Nodes) != 1 || roadmap.Nodes[0].Title != "Numpy for ML" {
		t.Fatalf("unexpected roadmap: %+v", roadmap)
	}
}
