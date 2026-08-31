package application_test

import (
	"context"
	"testing"

	assessmentDomain "github.com/hcl-backend/services/api-go/internal/assessment/domain"
	"github.com/hcl-backend/services/api-go/internal/progress/application"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
	"github.com/jackc/pgx/v5"
)

type mockProgressRepo struct {
	competent  int
	inProgress int
}

func (m *mockProgressRepo) RecordEvidence(ctx context.Context, tx pgx.Tx, evidence *domain.Evidence) error {
	return nil
}

func (m *mockProgressRepo) RecordEngagement(ctx context.Context, event *domain.EngagementEvent) error {
	return nil
}

func (m *mockProgressRepo) GetCompetentCount(ctx context.Context, learnerID string) (int, error) {
	return m.competent, nil
}

func (m *mockProgressRepo) GetInProgressCount(ctx context.Context, learnerID string) (int, error) {
	return m.inProgress, nil
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

func TestRecordEvidenceUseCase(t *testing.T) {
	repo := &mockProgressRepo{}
	compSvc := &mockCompetencyService{}
	txMgr := &mockTxManager{}

	uc := application.NewRecordEvidenceUseCase(txMgr, repo, compSvc)

	evidence := &assessmentDomain.Evidence{
		ID:        "ev-1",
		LearnerID: "l-1",
		ConceptID: "c-1",
		Result:    "competent",
	}

	err := uc.RecordEvidence(context.Background(), evidence)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !compSvc.called {
		t.Fatal("expected competency service to be called")
	}
}
