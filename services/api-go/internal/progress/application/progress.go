package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

// RecordEvidenceUseCase is the authoritative evidence pipeline: it writes the
// evidence row, transitions the learner's competency state, and syncs the
// roadmap node state — all inside one transaction.
type RecordEvidenceUseCase struct {
	txManager     TxManager
	repo          ProgressRepository
	competencySvc CompetencyService
}

func NewRecordEvidenceUseCase(txManager TxManager, repo ProgressRepository, competencySvc CompetencyService) *RecordEvidenceUseCase {
	return &RecordEvidenceUseCase{txManager: txManager, repo: repo, competencySvc: competencySvc}
}

// RecordEvidence inserts an evidence row and updates downstream state. It
// returns the deterministic competency state that resulted.
func (uc *RecordEvidenceUseCase) RecordEvidence(ctx context.Context, evidence *domain.Evidence) (string, error) {
	if evidence.LearnerID == "" || evidence.ConceptID == "" {
		return "", domain.ErrEvidenceFailed
	}
	if evidence.Result != "competent" && evidence.Result != "weak" && evidence.Result != "inconclusive" {
		return "", domain.ErrEvidenceFailed
	}

	evidence.ID = uuid.New().String()
	evidence.CreatedAt = time.Now().UTC()

	var resultState string
	err := uc.txManager.Do(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.repo.RecordEvidence(ctx, tx, evidence); err != nil {
			return err
		}
		if err := uc.competencySvc.UpdateWithEvidence(ctx, tx, evidence.LearnerID, evidence.ConceptID, evidence.Result, evidence.ID); err != nil {
			return err
		}
		if err := uc.repo.SyncPathItemState(ctx, tx, evidence.LearnerID, evidence.ConceptID, StateForResult(evidence.Result)); err != nil {
			return err
		}
		resultState = StateForResult(evidence.Result)
		return nil
	})
	if err != nil {
		return "", err
	}
	return resultState, nil
}

func StateForResult(result string) string {
	switch result {
	case "competent":
		return "competent"
	case "weak":
		return "weak_evidence"
	default:
		return "in_progress"
	}
}

// RecordEngagementUseCase records learner engagement events against path items.
type RecordEngagementUseCase struct {
	repo ProgressRepository
}

func NewRecordEngagementUseCase(repo ProgressRepository) *RecordEngagementUseCase {
	return &RecordEngagementUseCase{repo: repo}
}

func (uc *RecordEngagementUseCase) RecordEngagement(ctx context.Context, learnerID, conceptID, action string) error {
	if learnerID == "" || conceptID == "" || action == "" {
		return domain.ErrInvalidEvent
	}
	if action != "resource_opened" && action != "marked_reviewed" {
		return domain.ErrInvalidEvent
	}
	return uc.repo.RecordEngagement(ctx, &domain.EngagementEvent{
		ID:        uuid.New().String(),
		LearnerID: learnerID,
		ConceptID: conceptID,
		EventType: action,
		Timestamp: time.Now().UTC(),
	})
}

// GetProgressSummaryUseCase builds the tracking dashboard payload.
type GetProgressSummaryUseCase struct {
	repo  ProgressRepository
	goals GoalService
}

func NewGetProgressSummaryUseCase(repo ProgressRepository, goals GoalService) *GetProgressSummaryUseCase {
	return &GetProgressSummaryUseCase{repo: repo, goals: goals}
}

func (uc *GetProgressSummaryUseCase) Execute(ctx context.Context, learnerID string) (*domain.Summary, error) {
	_, _, structureID, err := uc.goals.ActiveStructureMeta(ctx, learnerID)
	if err != nil {
		return nil, err
	}
	payload, err := uc.repo.Summary(ctx, learnerID, structureID)
	if err != nil {
		return nil, err
	}
	overall := 0
	if payload.TotalConcepts > 0 {
		overall = payload.Competent * 100 / payload.TotalConcepts
	}
	return &domain.Summary{
		OverallCompletionPercentage: overall,
		TotalConcepts:               payload.TotalConcepts,
		CompletedConcepts:           payload.Competent,
		ActiveRemediations:          payload.Remediations,
		CompetencyBreakdown:         payload.Breakdown,
		ActivityData:                payload.ActivityData,
	}, nil
}

// GetGoalCompletionSummaryUseCase produces the goal completion certificate view.
type GetGoalCompletionSummaryUseCase struct {
	repo  ProgressRepository
	goals GoalService
}

func NewGetGoalCompletionSummaryUseCase(repo ProgressRepository, goals GoalService) *GetGoalCompletionSummaryUseCase {
	return &GetGoalCompletionSummaryUseCase{repo: repo, goals: goals}
}

func (uc *GetGoalCompletionSummaryUseCase) Execute(ctx context.Context, learnerID string) (*domain.GoalCompletionSummary, error) {
	goalID, goalTitle, structureID, err := uc.goals.ActiveStructureMeta(ctx, learnerID)
	if err != nil {
		return nil, err
	}
	names, err := uc.repo.CompetentConceptNames(ctx, learnerID, structureID)
	if err != nil {
		return nil, err
	}
	return &domain.GoalCompletionSummary{
		GoalID:              goalID,
		GoalTitle:           goalTitle,
		CompletionDate:      time.Now().UTC().Format("2006-01-02"),
		TotalSkillsVerified: len(names),
		MasteryProofList:    names,
	}, nil
}
