package application

import (
	"context"
	"time"

	"github.com/hcl-backend/services/api-go/internal/projects/domain"
)

type SubmitProjectUseCase struct {
	assessmentSvc AssessmentService
	storageSvc    StorageService
	evidenceSvc   EvidenceService
}

func NewSubmitProjectUseCase(
	assessmentSvc AssessmentService,
	storageSvc StorageService,
	evidenceSvc EvidenceService,
) *SubmitProjectUseCase {
	return &SubmitProjectUseCase{
		assessmentSvc: assessmentSvc,
		storageSvc:    storageSvc,
		evidenceSvc:   evidenceSvc,
	}
}

func (uc *SubmitProjectUseCase) Execute(ctx context.Context, learnerID, conceptID string, metadata domain.SubmissionMetadata) error {
	// 1. Verify project exists
	_, err := uc.assessmentSvc.GetProjectDefinition(ctx, conceptID)
	if err != nil {
		return domain.ErrNoProjectForConcept
	}

	// 2. Validate artifact
	if err := uc.storageSvc.ValidateArtifactReference(ctx, metadata.ArtifactReference); err != nil {
		return domain.ErrInvalidArtifact
	}

	// 3. Record evidence
	submission := &domain.ProjectSubmission{
		LearnerID:   learnerID,
		ConceptID:   conceptID,
		Metadata:    metadata,
		SubmittedAt: time.Now(),
	}

	if err := uc.evidenceSvc.RecordProjectSubmission(ctx, submission); err != nil {
		return domain.ErrSubmissionFailed
	}

	return nil
}
