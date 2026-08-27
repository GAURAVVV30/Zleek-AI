package application_test

import (
	"context"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/competency/application"
	"github.com/hcl-backend/services/api-go/internal/competency/domain"
	"github.com/jackc/pgx/v5"
)

type mockCompetencyRepo struct {
	record  *domain.CompetencyRecord
	history []domain.CompetencyHistory
}

func (m *mockCompetencyRepo) GetByLearnerAndConcept(ctx context.Context, learnerID, conceptID string) (*domain.CompetencyRecord, error) {
	if m.record != nil {
		return m.record, nil
	}
	return nil, domain.ErrCompetencyNotFound
}

func (m *mockCompetencyRepo) GetHistoryByLearnerAndConcept(ctx context.Context, learnerID, conceptID string) ([]domain.CompetencyHistory, error) {
	return m.history, nil
}

func (m *mockCompetencyRepo) UpsertCompetency(ctx context.Context, tx pgx.Tx, record *domain.CompetencyRecord) error {
	m.record = record
	return nil
}

func (m *mockCompetencyRepo) AppendHistory(ctx context.Context, tx pgx.Tx, history *domain.CompetencyHistory) error {
	m.history = append(m.history, *history)
	return nil
}

func TestUpdateCompetencyUseCase(t *testing.T) {
	repo := &mockCompetencyRepo{}
	uc := application.NewUpdateCompetencyUseCase(repo)

	// Valid Evidence transition to competent
	err := uc.UpdateWithEvidence(context.Background(), nil, "l-1", "c-1", "competent", "ev-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.record == nil || repo.record.State != domain.StateCompetent {
		t.Fatalf("expected state competent")
	}
	if len(repo.history) != 1 {
		t.Fatalf("expected 1 history record")
	}

	// Invalid Evidence transition
	err = uc.UpdateWithEvidence(context.Background(), nil, "l-1", "c-1", "fake-result", "ev-2")
	if err != domain.ErrInvalidState {
		t.Fatalf("expected ErrInvalidState, got %v", err)
	}
}
