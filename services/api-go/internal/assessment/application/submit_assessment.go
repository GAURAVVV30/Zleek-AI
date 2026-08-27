package application

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type SubmitAssessmentUseCase struct {
	repo            AssessmentRepository
	conceptService  ConceptService
	aiClient        AIClient
	evidenceService EvidenceService
}

func NewSubmitAssessmentUseCase(
	repo AssessmentRepository,
	conceptService ConceptService,
	aiClient AIClient,
	evidenceService EvidenceService,
) *SubmitAssessmentUseCase {
	return &SubmitAssessmentUseCase{
		repo:            repo,
		conceptService:  conceptService,
		aiClient:        aiClient,
		evidenceService: evidenceService,
	}
}

func (uc *SubmitAssessmentUseCase) Execute(ctx context.Context, learnerID, conceptID string, submissionData json.RawMessage) (*domain.EvaluationResult, error) {
	if err := uc.conceptService.ValidateConcept(ctx, conceptID); err != nil {
		return nil, domain.ErrConceptNotFound
	}

	def, err := uc.repo.GetDefinitionByConceptID(ctx, conceptID)
	if err != nil {
		return nil, domain.ErrAssessmentNotFound
	}

	if len(submissionData) == 0 {
		return nil, domain.ErrInvalidSubmission
	}

	evalResult, err := uc.aiClient.Evaluate(ctx, submissionData, def.Rubric)
	if err != nil {
		return nil, domain.ErrAIUnavailable
	}

	// Validate AI output
	if evalResult.Result != "competent" && evalResult.Result != "weak" && evalResult.Result != "inconclusive" {
		return nil, domain.ErrInvalidAIResult
	}

	if evalResult.Score < 0 || evalResult.Score > 100 || evalResult.Confidence < 0 || evalResult.Confidence > 1 {
		return nil, domain.ErrInvalidAIResult
	}

	evidence := &domain.Evidence{
		ID:                     uuid.New().String(),
		LearnerID:              learnerID,
		ConceptID:              conceptID,
		AssessmentDefinitionID: def.ID,
		SubmissionData:         submissionData,
		Score:                  evalResult.Score,
		Confidence:             evalResult.Confidence,
		EvaluatorType:          "ai",
		Result:                 evalResult.Result,
	}

	if err := uc.evidenceService.RecordEvidence(ctx, evidence); err != nil {
		return nil, err
	}

	return evalResult, nil
}
