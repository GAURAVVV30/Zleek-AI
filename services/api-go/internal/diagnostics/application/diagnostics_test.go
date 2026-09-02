package application_test

import (
	"context"
	"testing"

	"github.com/hcl-backend/services/api-go/internal/diagnostics/application"
	"github.com/hcl-backend/services/api-go/internal/diagnostics/domain"
)

type mockStore struct {
	sessions map[string]*domain.Session
}

func newMockStore() *mockStore {
	return &mockStore{sessions: map[string]*domain.Session{}}
}

func (m *mockStore) Create(ctx context.Context, s *domain.Session) error {
	m.sessions[s.SessionID] = s
	return nil
}

func (m *mockStore) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	s, ok := m.sessions[sessionID]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return s, nil
}

func (m *mockStore) Save(ctx context.Context, s *domain.Session) error {
	m.sessions[s.SessionID] = s
	return nil
}

type mockGoals struct{}

func (m *mockGoals) ActiveStructure(ctx context.Context, learnerID string) (string, string, error) {
	return "Become a ML Engineer", "ks-1", nil
}

type mockResolver struct{}

func (m *mockResolver) StructureDomainSlug(ctx context.Context, structureID string) (string, error) {
	return "machine_learning", nil
}

type mockGraph struct{}

func (m *mockGraph) TopoConcepts(ctx context.Context, domainSlug string) ([]application.NodeRef, error) {
	return []application.NodeRef{
		{NodeID: "ml_01", Name: "Programming Fundamentals"},
		{NodeID: "ml_02", Name: "Numpy"},
		{NodeID: "ml_03", Name: "Statistics"},
	}, nil
}

func (m *mockGraph) GetResources(domainSlug, nodeID string) []string {
	return []string{"RAG prerequisite context for " + nodeID}
}

type mockProfile struct{}
func (m *mockProfile) GetPriorExperience(ctx context.Context, learnerID string) (string, error) {
	return "Beginner", nil
}
func (m *mockProfile) GetRole(ctx context.Context, learnerID string) (string, error) {
	return "machine_learning", nil
}

type mockLLM struct{}
func (m *mockLLM) GenerateQuestionPrompt(ctx context.Context, role, priorLevel, conceptName, ragContext string) (*application.QuestionData, error) {
	return &application.QuestionData{
		Prompt:        "Mock question prompt for " + conceptName,
		Options:       []string{
			"I haven't touched this topic yet",
			"I've seen it but can't apply it on my own",
			"I can work through it with guidance",
			"I can apply it independently",
			"I could teach this topic to others",
		},
		CorrectOption: 0,
	}, nil
}
func (m *mockLLM) GenerateWeakAreasExplanation(ctx context.Context, role, priorLevel string, gaps []string, ragContext string) (string, error) {
	return "Mock weak areas explanation", nil
}

type mockCompetency struct{}
func (m *mockCompetency) SaveBaseline(ctx context.Context, learnerID, nodeID, state string) error {
	return nil
}

func TestDiagnosticFlow(t *testing.T) {
	store := newMockStore()
	start := application.NewStartDiagnosticUseCase(store, &mockGoals{}, &mockResolver{}, &mockGraph{}, &mockProfile{}, &mockLLM{})

	res, err := start.Execute(context.Background(), "l-1")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if res.TotalQuestions != 3 {
		t.Fatalf("expected 3 questions, got %d", res.TotalQuestions)
	}
	if res.FirstQuestion == nil || len(res.FirstQuestion.Options) != 3 {
		t.Fatalf("unexpected first question: %+v", res.FirstQuestion)
	}

	answer := application.NewAnswerDiagnosticUseCase(store)
	step, err := answer.Execute(context.Background(), "l-1", res.SessionID, "ml_01", "opt_1")
	if err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if step.IsComplete || step.NextQuestion == nil {
		t.Fatalf("expected a next question after first answer")
	}

	_, _ = answer.Execute(context.Background(), "l-1", res.SessionID, "ml_02", "opt_1")
	last, err := answer.Execute(context.Background(), "l-1", res.SessionID, "ml_03", "opt_3")
	if err != nil {
		t.Fatalf("answer 3: %v", err)
	}
	if !last.IsComplete {
		t.Fatal("expected completion on last answer")
	}

	resultsUseCase := application.NewResultsDiagnosticUseCase(store, &mockGoals{}, &mockResolver{}, &mockProfile{}, &mockGraph{}, &mockLLM{}, &mockCompetency{})
	results, err := resultsUseCase.Execute(context.Background(), "l-1", res.SessionID)
	if err != nil {
		t.Fatalf("results: %v", err)
	}
	if len(results.ConceptCoverage) != 3 {
		t.Fatalf("expected 3 coverage rows, got %d", len(results.ConceptCoverage))
	}
	if results.ConceptCoverage[2].Status != "gap" {
		t.Fatalf("expected low familiarity to be a gap, got %s", results.ConceptCoverage[2].Status)
	}
	if len(results.TopGaps) != 1 {
		t.Fatalf("expected 1 top gap, got %v", results.TopGaps)
	}
}

func TestDiagnosticAnswerValidation(t *testing.T) {
	store := newMockStore()
	start := application.NewStartDiagnosticUseCase(store, &mockGoals{}, &mockResolver{}, &mockGraph{}, &mockProfile{}, &mockLLM{})
	res, err := start.Execute(context.Background(), "l-1")
	if err != nil {
		t.Fatal(err)
	}
	answer := application.NewAnswerDiagnosticUseCase(store)
	if _, err := answer.Execute(context.Background(), "other-learner", res.SessionID, "ml_01", "opt_familiar_4"); err != domain.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound for wrong learner, got %v", err)
	}
	if _, err := answer.Execute(context.Background(), "l-1", res.SessionID, "bogus", "opt_familiar_4"); err != domain.ErrInvalidAnswer {
		t.Fatalf("expected ErrInvalidAnswer for unknown concept, got %v", err)
	}
}
