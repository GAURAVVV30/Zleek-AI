package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/assessment/application"
	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
)

type mockAssessmentRepo struct {
	def   *domain.AssessmentDefinition
	items []domain.AssessmentItem
	saved bool
}

func (m *mockAssessmentRepo) GetDefinitionByConceptID(ctx context.Context, conceptID string) (*domain.AssessmentDefinition, error) {
	if m.def != nil {
		return m.def, nil
	}
	return nil, domain.ErrAssessmentNotFound
}

func (m *mockAssessmentRepo) GetItemsByDefinitionID(ctx context.Context, definitionID string) ([]domain.AssessmentItem, error) {
	if m.items != nil {
		return m.items, nil
	}
	return nil, nil
}

func (m *mockAssessmentRepo) SaveDefinition(ctx context.Context, def *domain.AssessmentDefinition, items []domain.AssessmentItem) error {
	m.saved = true
	m.def = def
	m.items = items
	return nil
}

type mockConceptService struct{}

func (m *mockConceptService) ValidateConcept(ctx context.Context, conceptID string) error {
	if conceptID == "invalid-concept" {
		return domain.ErrConceptNotFound
	}
	return nil
}

func (m *mockConceptService) ConceptName(ctx context.Context, conceptID string) (string, error) {
	return "Pandas", nil
}

func (m *mockConceptService) CoreConceptNames(ctx context.Context, conceptID string) ([]string, error) {
	return []string{"DataFrames", "Series", "GroupBy"}, nil
}

func (m *mockConceptService) ConceptDomain(ctx context.Context, conceptID string) (string, error) {
	return "data-science", nil
}

type mockAIClient struct {
	result *domain.EvaluationResult
	err    error
}

func (m *mockAIClient) Evaluate(ctx context.Context, conceptID, domainID string, submission json.RawMessage) (*domain.EvaluationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &domain.EvaluationResult{Score: 85, Confidence: 0.95, Result: "competent", Feedback: "Solid answer."}, nil
}

type mockEvidenceService struct {
	called int
}

func (m *mockEvidenceService) RecordEvidence(ctx context.Context, evidence *domain.Evidence) (string, error) {
	m.called++
	return "competent", nil
}

func TestGetAssessmentUseCase_GeneratesQuiz(t *testing.T) {
	repo := &mockAssessmentRepo{}
	uc := application.NewGetAssessmentUseCase(repo, &mockConceptService{})

	quiz, err := uc.Execute(context.Background(), "pandas")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if quiz.ConceptID != "pandas" {
		t.Fatalf("wrong concept id: %q", quiz.ConceptID)
	}
	if quiz.ConceptTitle != "Pandas" {
		t.Fatalf("wrong title: %q", quiz.ConceptTitle)
	}
	if len(quiz.Questions) != 5 {
		t.Fatalf("expected 5 generated questions, got %d", len(quiz.Questions))
	}
	if !repo.saved {
		t.Fatal("expected definition to be saved on first access")
	}
	for _, q := range quiz.Questions {
		if len(q.Options) != 4 {
			t.Fatalf("expected 4 options per question, got %d", len(q.Options))
		}
		for _, o := range q.Options {
			if o.ID == "" || o.Text == "" {
				t.Fatalf("option missing id or text: %+v", o)
			}
		}
	}
}

func TestGetAssessmentUseCase_Deterministic(t *testing.T) {
	repo1 := &mockAssessmentRepo{}
	uc1 := application.NewGetAssessmentUseCase(repo1, &mockConceptService{})
	q1, err := uc1.Execute(context.Background(), "ml_01")
	if err != nil {
		t.Fatal(err)
	}

	repo2 := &mockAssessmentRepo{}
	uc2 := application.NewGetAssessmentUseCase(repo2, &mockConceptService{})
	q2, err := uc2.Execute(context.Background(), "ml_01")
	if err != nil {
		t.Fatal(err)
	}

	for i := range q1.Questions {
		if q1.Questions[i].Prompt != q2.Questions[i].Prompt {
			t.Fatalf("prompt mismatch for question %d: %q vs %q", i, q1.Questions[i].Prompt, q2.Questions[i].Prompt)
		}
		for j := range q1.Questions[i].Options {
			if q1.Questions[i].Options[j].ID != q2.Questions[i].Options[j].ID {
				t.Fatalf("option mismatch for question %d offset %d", i, j)
			}
		}
	}
}

func TestSubmitAssessmentUseCase_ObjectivePass(t *testing.T) {
	repo := &mockAssessmentRepo{}
	uc := application.NewGetAssessmentUseCase(repo, &mockConceptService{})
	quiz, err := uc.Execute(context.Background(), "pandas")
	if err != nil {
		t.Fatal(err)
	}

	submit := application.NewSubmitAssessmentUseCase(repo, &mockConceptService{}, &mockAIClient{}, &mockEvidenceService{})
	answers := make([]domain.AnswerSubmission, 0, len(quiz.Questions))
	for _, q := range quiz.Questions {
		// intentionally select the FIRST option; correctness determined by key
		answers = append(answers, domain.AnswerSubmission{QuestionID: q.ID, SelectedOptionID: q.Options[0].ID})
	}
	payload, _ := json.Marshal(map[string]any{"answers": answers})
	ev := &mockEvidenceService{}
	submit = application.NewSubmitAssessmentUseCase(repo, &mockConceptService{}, &mockAIClient{}, ev)

	res, err := submit.Execute(context.Background(), "learner-1", "pandas", payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.ScorePercentage < 0 || res.ScorePercentage > 100 {
		t.Fatalf("score out of range: %d", res.ScorePercentage)
	}
	if ev.called != 1 {
		t.Fatalf("expected evidence recorded once, got %d", ev.called)
	}
}

func TestSubmitAssessmentUseCase_FreeTextHitsAI(t *testing.T) {
	repo := &mockAssessmentRepo{}
	uc := application.NewGetAssessmentUseCase(repo, &mockConceptService{})
	if _, err := uc.Execute(context.Background(), "ml_01"); err != nil {
		t.Fatal(err)
	}
	ai := &mockAIClient{}
	ev := &mockEvidenceService{}
	submit := application.NewSubmitAssessmentUseCase(repo, &mockConceptService{}, ai, ev)

	payload, _ := json.Marshal(map[string]any{"freeText": "Pandas is a Python library for tabular data."})
	res, err := submit.Execute(context.Background(), "learner-1", "ml_01", payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Passed {
		t.Fatalf("expected pass for good free-text answer, got %+v", res)
	}
	if ev.called != 1 {
		t.Fatalf("expected evidence recorded once, got %d", ev.called)
	}
}

func TestSubmitAssessmentUseCase_MissingAnswerValidation(t *testing.T) {
	repo := &mockAssessmentRepo{}
	submit := application.NewSubmitAssessmentUseCase(repo, &mockConceptService{}, &mockAIClient{}, &mockEvidenceService{})
	_, err := submit.Execute(context.Background(), "learner-1", "pandas", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected validation error for empty submission")
	}
}
