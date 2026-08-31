package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/hcl-backend/services/api-go/internal/diagnostics/domain"
)

const diagnosticQuestionLimit = 5

// StartDiagnosticUseCase begins a diagnostic session for the learner's active
// goal, targeting the first concepts of the goal's domain.
type StartDiagnosticUseCase struct {
	store    SessionStore
	goals    GoalService
	resolver StructureResolver
	graph    GraphService
	profile  ProfileService
	llm      LLMService
}

func NewStartDiagnosticUseCase(store SessionStore, goals GoalService, resolver StructureResolver, graph GraphService, profile ProfileService, llm LLMService) *StartDiagnosticUseCase {
	return &StartDiagnosticUseCase{store: store, goals: goals, resolver: resolver, graph: graph, profile: profile, llm: llm}
}

func (uc *StartDiagnosticUseCase) Execute(ctx context.Context, learnerID string) (*StartResponse, error) {
	_, structureID, err := uc.goals.ActiveStructure(ctx, learnerID)
	if err != nil {
		return nil, domain.ErrNoActiveGoal
	}
	slug, err := uc.resolver.StructureDomainSlug(ctx, structureID)
	if err != nil {
		return nil, domain.ErrNoActiveGoal
	}
	refs, err := uc.graph.TopoConcepts(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("resolve domain concepts: %w", err)
	}
	if len(refs) > diagnosticQuestionLimit {
		refs = refs[:diagnosticQuestionLimit]
	}
	if len(refs) == 0 {
		return nil, domain.ErrNoActiveGoal
	}

	concepts := make([]domain.Concept, 0, len(refs))
	for _, r := range refs {
		concepts = append(concepts, domain.Concept{NodeID: r.NodeID, Name: r.Name})
	}

	priorLevel, err := uc.profile.GetPriorExperience(ctx, learnerID)
	if err != nil {
		priorLevel = "Beginner"
	}

	role, err := uc.profile.GetRole(ctx, learnerID)
	if err != nil || role == "" {
		role = slug
	}

	prompts := make(map[string]string)
	questions := make([]domain.Question, 0, len(concepts))
	correctAnswers := make(map[string]string)

	for i, c := range concepts {
		ragResources := uc.graph.GetResources(slug, c.NodeID)
		ragContext := strings.Join(ragResources, "\n---\n")

		qData, err := uc.llm.GenerateQuestionPrompt(ctx, role, priorLevel, c.Name, ragContext)
		if err == nil && qData != nil {
			prompts[c.NodeID] = qData.Prompt
			opts := make([]domain.QuestionOption, 0, len(qData.Options))
			for idx, optText := range qData.Options {
				opts = append(opts, domain.QuestionOption{
					ID:   fmt.Sprintf("opt_familiar_%d", idx+1),
					Text: optText,
				})
			}
			questions = append(questions, domain.Question{
				QuestionID:     c.NodeID,
				QuestionNumber: i + 1,
				TotalQuestions: len(concepts),
				ConceptID:      c.NodeID,
				ConceptName:    c.Name,
				Prompt:         qData.Prompt,
				Options:        opts,
			})
			correctAnswers[c.NodeID] = fmt.Sprintf("opt_familiar_%d", qData.CorrectOption+1)
		} else {
			prompts[c.NodeID] = "How comfortable are you with " + c.Name + "?"
			opts := []domain.QuestionOption{
				{ID: "opt_familiar_1", Text: "Yes"},
				{ID: "opt_familiar_2", Text: "No"},
			}
			questions = append(questions, domain.Question{
				QuestionID:     c.NodeID,
				QuestionNumber: i + 1,
				TotalQuestions: len(concepts),
				ConceptID:      c.NodeID,
				ConceptName:    c.Name,
				Prompt:         "How comfortable are you with " + c.Name + "?",
				Options:        opts,
			})
			correctAnswers[c.NodeID] = "opt_familiar_1"
		}
	}

	session := &domain.Session{
		SessionID:      "diag_" + uuid.NewString(),
		LearnerID:      learnerID,
		Concepts:       concepts,
		Answers:        map[string]string{},
		Prompts:        prompts,
		CorrectAnswers: correctAnswers,
		Questions:      questions,
	}
	if err := uc.store.Create(ctx, session); err != nil {
		return nil, err
	}
	return &StartResponse{
		SessionID:      session.SessionID,
		FirstQuestion:  &session.Questions[0],
		TotalQuestions: len(session.Questions),
	}, nil
}

// AnswerDiagnosticUseCase records one answer and advances the session.
type AnswerDiagnosticUseCase struct {
	store SessionStore
}

func NewAnswerDiagnosticUseCase(store SessionStore) *AnswerDiagnosticUseCase {
	return &AnswerDiagnosticUseCase{store: store}
}

func (uc *AnswerDiagnosticUseCase) Execute(ctx context.Context, learnerID, sessionID, questionID, optionID string) (*domain.AnswerResponse, error) {
	session, err := uc.store.Get(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}
	if session.LearnerID != learnerID {
		return nil, domain.ErrSessionNotFound
	}
	if session.Completed {
		return nil, domain.ErrSessionComplete
	}
	if !validConcept(session, questionID) {
		return nil, domain.ErrInvalidAnswer
	}

	var currentQuestion *domain.Question
	for _, q := range session.Questions {
		if q.QuestionID == questionID {
			currentQuestion = &q
			break
		}
	}
	if currentQuestion != nil {
		isValidOption := false
		for _, opt := range currentQuestion.Options {
			if opt.ID == optionID {
				isValidOption = true
				break
			}
		}
		if !isValidOption {
			return nil, domain.ErrInvalidAnswer
		}
	} else {
		if coverageFor(optionID) == 10 && optionID != "opt_familiar_1" {
			return nil, domain.ErrInvalidAnswer
		}
	}

	session.Answers[questionID] = optionID

	idx := indexOfConcept(session, questionID)
	var next *domain.Question
	if len(session.Questions) > 0 {
		if idx+1 < len(session.Questions) {
			next = &session.Questions[idx+1]
		} else {
			session.Completed = true
		}
	} else {
		if idx+1 < len(session.Concepts) {
			q := BuildQuestionsWithPrompts(session.Concepts, session.Prompts)[idx+1]
			next = &q
		} else {
			session.Completed = true
		}
	}
	if err := uc.store.Save(ctx, session); err != nil {
		return nil, err
	}
	return &domain.AnswerResponse{IsComplete: session.Completed, NextQuestion: next}, nil
}

// ResultsDiagnosticUseCase computes the baseline from a completed session.
type ResultsDiagnosticUseCase struct {
	store      SessionStore
	goals      GoalService
	resolver   StructureResolver
	profile    ProfileService
	graph      GraphService
	llm        LLMService
	competency CompetencyService
}

func NewResultsDiagnosticUseCase(store SessionStore, goals GoalService, resolver StructureResolver, profile ProfileService, graph GraphService, llm LLMService, competency CompetencyService) *ResultsDiagnosticUseCase {
	return &ResultsDiagnosticUseCase{store: store, goals: goals, resolver: resolver, profile: profile, graph: graph, llm: llm, competency: competency}
}

func (uc *ResultsDiagnosticUseCase) Execute(ctx context.Context, learnerID, sessionID string) (*domain.BaselineResults, error) {
	session, err := uc.store.Get(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}
	if session.LearnerID != learnerID {
		return nil, domain.ErrSessionNotFound
	}

	res := computeResults(session)

	// Save baseline competencies to database based on answers
	for _, row := range res.ConceptCoverage {
		compState := "not_started"
		switch row.Status {
		case "gap":
			compState = "weak_evidence"
		case "in_progress":
			compState = "in_progress"
		case "competent":
			compState = "competent"
		}
		_ = uc.competency.SaveBaseline(ctx, learnerID, row.ConceptID, compState)
	}

	priorLevel, err := uc.profile.GetPriorExperience(ctx, learnerID)
	if err != nil {
		priorLevel = "Beginner"
	}

	role, err := uc.profile.GetRole(ctx, learnerID)
	if err != nil || role == "" {
		_, structureID, err := uc.goals.ActiveStructure(ctx, learnerID)
		if err == nil {
			role, _ = uc.resolver.StructureDomainSlug(ctx, structureID)
		}
	}
	if role == "" {
		role = "machine_learning"
	}

	var ragContextParts []string
	for _, gap := range res.TopGaps {
		var nodeID string
		for _, c := range session.Concepts {
			if c.Name == gap {
				nodeID = c.NodeID
				break
			}
		}
		if nodeID != "" {
			ragResources := uc.graph.GetResources(role, nodeID)
			if len(ragResources) > 0 {
				ragContextParts = append(ragContextParts, fmt.Sprintf("Concept: %s\nContent:\n%s", gap, strings.Join(ragResources, "\n")))
			}
		}
	}
	ragContext := strings.Join(ragContextParts, "\n---\n")

	explanation, err := uc.llm.GenerateWeakAreasExplanation(ctx, role, priorLevel, res.TopGaps, ragContext)
	if err == nil {
		res.Explanation = explanation
	} else {
		res.Explanation = "Based on your diagnostic assessment, we have identified areas for improvement to tailor your custom learning path."
	}

	return res, nil
}

type StartResponse struct {
	SessionID      string           `json:"sessionId"`
	FirstQuestion  *domain.Question `json:"firstQuestion"`
	TotalQuestions int              `json:"totalQuestions"`
}

func validConcept(s *domain.Session, nodeID string) bool {
	for _, c := range s.Concepts {
		if c.NodeID == nodeID {
			return true
		}
	}
	return false
}

func indexOfConcept(s *domain.Session, nodeID string) int {
	for i, c := range s.Concepts {
		if c.NodeID == nodeID {
			return i
		}
	}
	return -1
}
