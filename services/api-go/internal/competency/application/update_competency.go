package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hcl-backend/services/api-go/internal/competency/domain"
	"github.com/jackc/pgx/v5"
)

type UpdateCompetencyUseCase struct {
	repo CompetencyRepository
}

func NewUpdateCompetencyUseCase(repo CompetencyRepository) *UpdateCompetencyUseCase {
	return &UpdateCompetencyUseCase{
		repo: repo,
	}
}

// UpdateWithEvidence is called by the Progress module within an existing transaction
func (uc *UpdateCompetencyUseCase) UpdateWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, conceptID, result, evidenceID string) error {
	if result == "" || evidenceID == "" {
		return domain.ErrInvalidEvidence
	}

	current, err := uc.repo.GetByLearnerAndConcept(ctx, learnerID, conceptID)

	var prevState *string
	var newState domain.CompetencyState

	if err == nil && current != nil {
		s := string(current.State)
		prevState = &s
	}

	// Map Evidence Result to Competency State deterministically
	switch result {
	case "competent":
		newState = domain.StateCompetent
	case "weak":
		newState = domain.StateWeakEvidence
	case "inconclusive":
		newState = domain.StateInProgress
	default:
		return domain.ErrInvalidState
	}

	if current != nil {
		current.State = newState
		current.LastEvidenceID = &evidenceID
		current.UpdatedAt = time.Now()
	} else {
		current = &domain.CompetencyRecord{
			LearnerID:      learnerID,
			ConceptID:      conceptID,
			State:          newState,
			LastEvidenceID: &evidenceID,
			UpdatedAt:      time.Now(),
		}
	}

	if err := uc.repo.UpsertCompetency(ctx, tx, current); err != nil {
		return err
	}

	history := &domain.CompetencyHistory{
		ID:            uuid.New().String(),
		LearnerID:     learnerID,
		ConceptID:     conceptID,
		PreviousState: prevState,
		NewState:      string(newState),
		EvidenceID:    evidenceID,
		ChangedAt:     time.Now(),
	}

	if err := uc.repo.AppendHistory(ctx, tx, history); err != nil {
		return err
	}

	return nil
}
