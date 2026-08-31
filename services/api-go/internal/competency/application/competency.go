package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/competency/domain"
)

// ConceptService validates a concept by its public roadmap node id.
type ConceptService interface {
	ValidateConcept(ctx context.Context, conceptID string) error
}

type CompetencyRepository interface {
	GetByConceptNodeID(ctx context.Context, learnerID, nodeID string) (*domain.CompetencyRecord, error)
	GetHistoryByConceptNodeID(ctx context.Context, learnerID, nodeID string) ([]domain.CompetencyHistory, error)
	ListByLearner(ctx context.Context, learnerID string) ([]domain.CompetencyRecord, error)
	UpsertWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, nodeID, state, evidenceID string) (previousState string, err error)
	AppendHistory(ctx context.Context, tx pgx.Tx, h *HistoryRow) error
	CreateBaseline(ctx context.Context, learnerID, nodeID, state string) error
}

// HistoryRow is a plain append row (concept addressed by public node id).
type HistoryRow struct {
	HistoryID     string
	LearnerID     string
	NodeID        string
	PreviousState string
	NewState      string
	EvidenceID    string
	ChangedAt     time.Time
}

// StateForResult maps evidence results to competency states (deterministic).
func StateForResult(result string) string {
	switch result {
	case "competent":
		return string(domain.StateCompetent)
	case "weak":
		return string(domain.StateWeak)
	default:
		return string(domain.StateInProgress)
	}
}

// UpdateCompetencyUseCase drives the deterministic evidence → state machine.
type UpdateCompetencyUseCase struct {
	repo CompetencyRepository
}

func NewUpdateCompetencyUseCase(repo CompetencyRepository) *UpdateCompetencyUseCase {
	return &UpdateCompetencyUseCase{repo: repo}
}

// UpdateWithEvidence records a state transition for evidence within the caller
// transaction.
func (uc *UpdateCompetencyUseCase) UpdateWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, nodeID, result, evidenceID string) error {
	if result != "competent" && result != "weak" && result != "inconclusive" {
		return domain.ErrInvalidState
	}
	if evidenceID == "" {
		return domain.ErrInvalidEvidence
	}
	state := StateForResult(result)
	previous, err := uc.repo.UpsertWithEvidence(ctx, tx, learnerID, nodeID, state, evidenceID)
	if err != nil {
		return err
	}
	if previous == "" {
		previous = string(domain.StateNotStarted)
	}
	if previous == state {
		return nil
	}
	return uc.repo.AppendHistory(ctx, tx, &HistoryRow{
		HistoryID:     uuid.New().String(),
		LearnerID:     learnerID,
		NodeID:        nodeID,
		PreviousState: previous,
		NewState:      state,
		EvidenceID:    evidenceID,
		ChangedAt:     time.Now().UTC(),
	})
}

// UnknownStateFor writes a baseline competency when a learner has not provided
// evidence yet (e.g. diagnostic gaps).
func (uc *UpdateCompetencyUseCase) UnknownStateFor(ctx context.Context, learnerID, nodeID string, mark string) error {
	state := string(domain.StateInProgress)
	if mark == "gap" {
		state = string(domain.StateWeak)
	}
	return uc.repo.CreateBaseline(ctx, learnerID, nodeID, state)
}

// GetCompetencyDetailUseCase returns the learner's competency table.
type GetCompetencyDetailUseCase struct {
	repo CompetencyRepository
}

func NewGetCompetencyDetailUseCase(repo CompetencyRepository, _ ConceptService) *GetCompetencyDetailUseCase {
	return &GetCompetencyDetailUseCase{repo: repo}
}

func (uc *GetCompetencyDetailUseCase) Execute(ctx context.Context, learnerID string) ([]domain.CompetencyRecord, error) {
	return uc.repo.ListByLearner(ctx, learnerID)
}

// GetCompetencyHistoryUseCase returns evidence history for a concept.
type GetCompetencyHistoryUseCase struct {
	repo CompetencyRepository
}

func NewGetCompetencyHistoryUseCase(repo CompetencyRepository, _ ConceptService) *GetCompetencyHistoryUseCase {
	return &GetCompetencyHistoryUseCase{repo: repo}
}

func (uc *GetCompetencyHistoryUseCase) Execute(ctx context.Context, learnerID, conceptID string) ([]domain.CompetencyHistory, error) {
	return uc.repo.GetHistoryByConceptNodeID(ctx, learnerID, conceptID)
}
