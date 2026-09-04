package application_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/progress/application"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type mockProgressRepo struct {
	evidence         *domain.Evidence
	synced           int
	summary          *application.SummaryPayload
	names            []string
	totalModules     int
	completedModules int
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

func (m *mockProgressRepo) GetCompletionBadgeStatus(ctx context.Context, learnerID, structureID string) (int, int, error) {
	return m.totalModules, m.completedModules, nil
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

func TestGetCompletionBadgeUseCase(t *testing.T) {
	tests := []struct {
		name             string
		goalTitle        string
		totalModules     int
		completedModules int
		expectedEligible bool
		expectedRole     string
	}{
		{
			name:             "CASE 1: 0 / N modules completed -> badge forbidden",
			goalTitle:        "I want to master AI Engineer from the ground up.",
			totalModules:     7,
			completedModules: 0,
			expectedEligible: false,
			expectedRole:     "AI ENGINEER",
		},
		{
			name:             "CASE 2: 1 / N completed -> badge forbidden",
			goalTitle:        "Become a Data Analyst",
			totalModules:     7,
			completedModules: 1,
			expectedEligible: false,
			expectedRole:     "DATA ANALYST",
		},
		{
			name:             "CASE 3: N-1 / N completed -> badge forbidden",
			goalTitle:        "Master Backend Engineer Track",
			totalModules:     7,
			completedModules: 6,
			expectedEligible: false,
			expectedRole:     "BACKEND ENGINEER",
		},
		{
			name:             "CASE 4: N / N completed -> badge allowed",
			goalTitle:        "I want to master AI Engineer from the ground up.",
			totalModules:     7,
			completedModules: 7,
			expectedEligible: true,
			expectedRole:     "AI ENGINEER",
		},
		{
			name:             "CASE 10 & 11: Dynamic Runtime Role Formatting for Mobile Engineer",
			goalTitle:        "Become a Mobile Engineer",
			totalModules:     5,
			completedModules: 5,
			expectedEligible: true,
			expectedRole:     "MOBILE ENGINEER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProgressRepo{
				totalModules:     tt.totalModules,
				completedModules: tt.completedModules,
			}
			goals := &mockGoalService{
				goalID:    "g-1",
				title:     tt.goalTitle,
				structure: "s-1",
			}

			useCase := application.NewGetCompletionBadgeUseCase(repo, goals)
			res, err := useCase.Execute(context.Background(), "learner-uuid-123")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if res.Eligible != tt.expectedEligible {
				t.Errorf("expected eligible=%v, got=%v", tt.expectedEligible, res.Eligible)
			}

			if res.Role != tt.expectedRole {
				t.Errorf("expected role=%q, got=%q", tt.expectedRole, res.Role)
			}

			if tt.expectedEligible {
				if res.Badge == nil || res.Badge.Title != tt.expectedRole || res.Badge.Status != "verified" {
					t.Errorf("expected verified badge details for role %q, got %+v", tt.expectedRole, res.Badge)
				}
			} else {
				if res.Message == "" {
					t.Errorf("expected explanatory rejection message when incomplete")
				}
			}
		})
	}
}
