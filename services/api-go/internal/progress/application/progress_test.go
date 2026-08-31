package application_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/progress/application"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type mockProgressRepo struct {
	evidence *domain.Evidence
	synced   int
	summary  *application.SummaryPayload
	names    []string
}

func (m *mockProgressRepo) RecordEvidence(ctx context.Context, tx pgx.Tx, evidence *domain.Evidence) error {
	m.evidence = evidence
	return nil
}

func (m *mockProgressRepo) RecordEngagement(ctx context.Context, event *domain.EngagementEvent) error {
	return nil
}

func (m *mockProgressRepo) SyncPathItemState(ctx context.Context, tx pgx.Tx, learnerID, conceptNodeID, state string) error {
	m.synced++
	return nil
}

func (m *mockProgressRepo) Summary(ctx context.Context, learnerID, structureID string) (*application.SummaryPayload, error) {
	return m.summary, nil
}

func (m *mockProgressRepo) CompetentConceptNames(ctx context.Context, learnerID, structureID string) ([]string, error) {
	return m.names, nil
}

type mockCompetencyService struct {
	called bool
}

func (m *mockCompetencyService) UpdateWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, conceptID, result, evidenceID string) error {
	m.called = true
	return nil
}

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

type mockGoalService struct {
	goalID    string
	title     string
	structure string
}

func (m *mockGoalService) ActiveStructureMeta(ctx context.Context, learnerID string) (string, string, string, error) {
	return m.goalID, m.title, m.structure, nil
}

func TestRecordEvidenceUseCase(t *testing.T) {
	repo := &mockProgressRepo{}
	compSvc := &mockCompetencyService{}
	txMgr := &mockTxManager{}

	uc := application.NewRecordEvidenceUseCase(txMgr, repo, compSvc)

	state, err := uc.RecordEvidence(context.Background(), &domain.Evidence{
		LearnerID: "l-1",
		ConceptID: "ml_01",
		Result:    "competent",
		Score:     90,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state != "competent" {
		t.Fatalf("expected state competent, got %q", state)
	}
	if !compSvc.called {
		t.Fatal("expected competency service to be called")
	}
	if repo.evidence == nil || repo.evidence.ID == "" {
		t.Fatal("expected evidence to be recorded with generated ID")
	}
}

func TestRecordEvidenceUseCase_InvalidResult(t *testing.T) {
	uc := application.NewRecordEvidenceUseCase(&mockTxManager{}, &mockProgressRepo{}, &mockCompetencyService{})
	if _, err := uc.RecordEvidence(context.Background(), &domain.Evidence{
		LearnerID: "l-1",
		ConceptID: "ml_01",
		Result:    "bogus",
	}); err == nil {
		t.Fatal("expected error for invalid result")
	}
}

func TestGetProgressSummaryUseCase(t *testing.T) {
	repo := &mockProgressRepo{
		summary: &application.SummaryPayload{
			TotalConcepts: 12,
			Competent:     6,
			Breakdown:     []domain.SummaryRow{{Domain: "Programming Fundamentals", Percentage: 100, Status: "Completed"}},
		},
	}
	goals := &mockGoalService{goalID: "g-1", title: "Become a Machine Learning Engineer", structure: "s-1"}
	uc := application.NewGetProgressSummaryUseCase(repo, goals)

	summary, err := uc.Execute(context.Background(), "l-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.OverallCompletionPercentage != 50 {
		t.Fatalf("expected 50%% overall, got %d", summary.OverallCompletionPercentage)
	}
	if len(summary.CompetencyBreakdown) != 1 {
		t.Fatalf("expected 1 breakdown row, got %d", len(summary.CompetencyBreakdown))
	}
}
