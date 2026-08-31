package application_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/competency/application"
	"github.com/hcl-backend/services/api-go/internal/competency/domain"
)

type mockCompetencyRepo struct {
	upserts  int
	previous string
	history  int
}

func (m *mockCompetencyRepo) GetByConceptNodeID(ctx context.Context, learnerID, nodeID string) (*domain.CompetencyRecord, error) {
	return nil, domain.ErrCompetencyNotFound
}

func (m *mockCompetencyRepo) GetHistoryByConceptNodeID(ctx context.Context, learnerID, nodeID string) ([]domain.CompetencyHistory, error) {
	return nil, nil
}

func (m *mockCompetencyRepo) ListByLearner(ctx context.Context, learnerID string) ([]domain.CompetencyRecord, error) {
	return nil, nil
}

func (m *mockCompetencyRepo) UpsertWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, nodeID, state, evidenceID string) (string, error) {
	m.upserts++
	return m.previous, nil
}

func (m *mockCompetencyRepo) AppendHistory(ctx context.Context, tx pgx.Tx, h *application.HistoryRow) error {
	m.history++
	return nil
}

func (m *mockCompetencyRepo) CreateBaseline(ctx context.Context, learnerID, nodeID, state string) error {
	return nil
}

func TestUpdateWithEvidence_Competent(t *testing.T) {
	repo := &mockCompetencyRepo{previous: "not_started"}
	uc := application.NewUpdateCompetencyUseCase(repo)

	err := uc.UpdateWithEvidence(context.Background(), nil, "l-1", "ml_01", "competent", "ev-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.upserts != 1 {
		t.Fatalf("expected 1 upsert, got %d", repo.upserts)
	}
	if repo.history != 1 {
		t.Fatalf("expected 1 history append, got %d", repo.history)
	}
}

func TestUpdateWithEvidence_NoOpWhenStateUnchanged(t *testing.T) {
	repo := &mockCompetencyRepo{previous: "competent"}
	uc := application.NewUpdateCompetencyUseCase(repo)

	if err := uc.UpdateWithEvidence(context.Background(), nil, "l-1", "ml_01", "competent", "ev-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.history != 0 {
		t.Fatalf("expected no history for unchanged state, got %d", repo.history)
	}
}

func TestUpdateWithEvidence_InvalidResult(t *testing.T) {
	repo := &mockCompetencyRepo{}
	uc := application.NewUpdateCompetencyUseCase(repo)

	if err := uc.UpdateWithEvidence(context.Background(), nil, "l-1", "ml_01", "bogus", "ev-1"); err == nil {
		t.Fatal("expected error for unknown result")
	}
}

func TestStateForResult(t *testing.T) {
	cases := map[string]string{
		"competent":    "competent",
		"weak":         "weak_evidence",
		"inconclusive": "in_progress",
	}
	for in, want := range cases {
		if got := application.StateForResult(in); got != want {
			t.Errorf("StateForResult(%q) = %q, want %q", in, got, want)
		}
	}
}
