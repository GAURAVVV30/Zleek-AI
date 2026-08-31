package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

const passThreshold = 70

// SubmitAssessmentUseCase grades a quiz submission (objective, deterministic)
// or a free-text answer (AI-evaluated) and records evidence.
type SubmitAssessmentUseCase struct {
	repo           AssessmentRepository
	conceptCatalog ConceptCatalog
	aiClient       AIClient
	evidence       EvidenceService
}

func NewSubmitAssessmentUseCase(
	repo AssessmentRepository,
	conceptCatalog ConceptCatalog,
	aiClient AIClient,
	evidence EvidenceService,
) *SubmitAssessmentUseCase {
	return &SubmitAssessmentUseCase{repo: repo, conceptCatalog: conceptCatalog, aiClient: aiClient, evidence: evidence}
}

func (uc *SubmitAssessmentUseCase) Execute(ctx context.Context, learnerID, conceptID string, submissionRaw json.RawMessage) (*domain.QuizResult, error) {
	if err := uc.conceptCatalog.ValidateConcept(ctx, conceptID); err != nil {
		return nil, domain.ErrConceptNotFound
	}
	if len(submissionRaw) == 0 {
		return nil, domain.ErrInvalidSubmission
	}
	if learnerID == "" {
		return nil, domain.ErrInvalidSubmission
	}

	def, items, err := ensureDefinition(ctx, uc.repo, uc.conceptCatalog, conceptID)
	if err != nil {
		return nil, err
	}

	var submission domain.QuizSubmission
	if err := json.Unmarshal(submissionRaw, &submission); err != nil {
		return nil, domain.ErrInvalidSubmission
	}

	if len(submission.Answers) == 0 && submission.FreeText == "" {
		return nil, domain.ErrInvalidSubmission
	}

	var result *domain.QuizResult
	if submission.FreeText != "" {
		result, err = uc.gradeFreeText(ctx, conceptID, def, submissionRaw)
	} else {
		result, err = uc.gradeObjective(def, items, submission)
	}
	if err != nil {
		return nil, err
	}

	newState, err := uc.evidence.RecordEvidence(ctx, &domain.Evidence{
		LearnerID:              learnerID,
		ConceptID:              conceptID,
		AssessmentDefinitionID: def.ID,
		SubmissionData:         submissionRaw,
		Score:                  float64(result.ScorePercentage),
		Confidence:             result.Confidence,
		EvaluatorType:          "ai",
		Result:                 result.Result,
	})
	if err != nil {
		return nil, err
	}
	result.NewCompetencyState = newState
	return result, nil
}

func (uc *SubmitAssessmentUseCase) gradeObjective(def *domain.AssessmentDefinition, items []domain.AssessmentItem, sub domain.QuizSubmission) (*domain.QuizResult, error) {
	if len(items) == 0 {
		return nil, domain.ErrInvalidSubmission
	}
	answerByID := map[string]domain.AssessmentItem{}
	for _, item := range items {
		answerByID[item.ID] = item
	}
	correct := 0
	for _, a := range sub.Answers {
		item, ok := answerByID[a.QuestionID]
		if !ok {
			continue
		}
		var key domain.QuizAnswerKey
		if json.Unmarshal(item.AnswerKey, &key) != nil {
			continue
		}
		if a.SelectedOptionID == key.Correct {
			correct++
		}
	}
	pct := correct * 100 / len(items)
	result := buildResult(pct)
	_ = def
	result.Score = float64(pct)
	result.Confidence = 1.0
	result.Result = "competent"
	if pct < passThreshold {
		result.Result = "weak"
	}
	result.Feedback = fmt.Sprintf("You answered %d of %d questions correctly.", correct, len(items))
	if pct >= passThreshold {
		result.Feedback += " The evaluator verified your understanding."
	} else {
		result.Feedback += " Review the core concepts and prerequisite topics before retrying."
	}
	return result, nil
}

func (uc *SubmitAssessmentUseCase) gradeFreeText(ctx context.Context, conceptID string, def *domain.AssessmentDefinition, submission json.RawMessage) (*domain.QuizResult, error) {
	domainID, err := uc.conceptCatalog.ConceptDomain(ctx, conceptID)
	if err != nil {
		return nil, domain.ErrConceptNotFound
	}
	eval, err := uc.aiClient.Evaluate(ctx, conceptID, domainID, submission)
	if err != nil {
		return nil, domain.ErrAIUnavailable
	}
	if eval.Result != "competent" && eval.Result != "weak" && eval.Result != "inconclusive" {
		return nil, domain.ErrInvalidAIResult
	}
	pct := int(eval.Score)
	if pct > 100 {
		pct = 100
	}
	if pct < 0 {
		pct = 0
	}
	result := buildResult(pct)
	result.Score = float64(pct)
	result.Confidence = eval.Confidence
	result.Result = eval.Result
	if result.Result == "inconclusive" {
		result.Result = "weak"
	}
	if eval.Feedback != "" {
		result.Feedback = eval.Feedback
	}
	return result, nil
}

func buildResult(pct int) *domain.QuizResult {
	passed := pct >= passThreshold
	return &domain.QuizResult{
		Passed:               passed,
		ScorePercentage:      pct,
		Feedback:             "Evidence recorded.",
		RemediationTriggered: !passed,
	}
}
