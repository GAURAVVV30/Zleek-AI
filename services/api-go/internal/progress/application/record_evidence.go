package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
	assessmentDomain "github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type RecordEvidenceUseCase struct {
	txManager         database.TxManager
	repo              ProgressRepository
	competencyService CompetencyService
}

func NewRecordEvidenceUseCase(txManager database.TxManager, repo ProgressRepository, compService CompetencyService) *RecordEvidenceUseCase {
	return &RecordEvidenceUseCase{
		txManager:         txManager,
		repo:              repo,
		competencyService: compService,
	}
}

// RecordEvidence translates the assessment-domain Evidence into progress-domain Evidence,
// saves it, and invokes the Competency boundary to update state in a single transaction.
func (uc *RecordEvidenceUseCase) RecordEvidence(ctx context.Context, asstEvidence *assessmentDomain.Evidence) error {
	evidence := &domain.Evidence{
		ID:                     asstEvidence.ID,
		LearnerID:              asstEvidence.LearnerID,
		ConceptID:              asstEvidence.ConceptID,
		AssessmentDefinitionID: asstEvidence.AssessmentDefinitionID,
		SubmissionData:         asstEvidence.SubmissionData,
		Score:                  asstEvidence.Score,
		Confidence:             asstEvidence.Confidence,
		EvaluatorType:          asstEvidence.EvaluatorType,
		Result:                 asstEvidence.Result,
		CreatedAt:              time.Now(),
	}

	return uc.txManager.Do(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Record evidence
		if err := uc.repo.RecordEvidence(ctx, tx, evidence); err != nil {
			return domain.ErrEvidenceFailed
		}

		// 2. Delegate competency update logic to the Competency boundary
		if err := uc.competencyService.UpdateWithEvidence(ctx, tx, evidence.LearnerID, evidence.ConceptID, evidence.Result, evidence.ID); err != nil {
			return err
		}

		return nil
	})
}
