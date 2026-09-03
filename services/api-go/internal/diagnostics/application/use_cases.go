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
		cleanID := strings.TrimSpace(r.NodeID)
		concepts = append(concepts, domain.Concept{NodeID: cleanID, Name: r.Name})
	}

	prefs, err := uc.profile.GetPreferences(ctx, learnerID)
	priorLevel := "Beginner"
	formatPreference := ""
	timeAvailability := ""
	if err == nil && prefs != nil {
		if prefs.PriorExperience != "" {
			priorLevel = prefs.PriorExperience
		}
		formatPreference = prefs.FormatPreference
		timeAvailability = prefs.TimeAvailability
	} else {
		if p, pErr := uc.profile.GetPriorExperience(ctx, learnerID); pErr == nil && p != "" {
			priorLevel = p
		}
	}

	role, err := uc.profile.GetRole(ctx, learnerID)
	if err != nil || role == "" {
		role = slug
	}

	prompts := make(map[string]string)
	questions := make([]domain.Question, 0, len(concepts))
	correctAnswers := make(map[string]string)

	for i, c := range concepts {
		cleanID := strings.TrimSpace(c.NodeID)
		ragResources := uc.graph.GetResources(slug, cleanID)
		ragContext := strings.Join(ragResources, "\n---\n")

		var qData *QuestionData
		var genErr error
		for attempt := 0; attempt < 3; attempt++ {
			qData, genErr = uc.llm.GenerateQuestionPrompt(ctx, role, priorLevel, formatPreference, timeAvailability, c.Name, ragContext)
			if genErr == nil && qData != nil && strings.TrimSpace(qData.Prompt) != "" && len(qData.Options) == 3 {
				break
			}
		}

		if qData == nil || strings.TrimSpace(qData.Prompt) == "" || len(qData.Options) != 3 {
			qData = buildFallbackQuestion(role, priorLevel, c.Name)
		}

		correctIdx := qData.CorrectOption
		if correctIdx < 0 || correctIdx >= 3 {
			correctIdx = 0
		}

		prompts[cleanID] = qData.Prompt
		opts := make([]domain.QuestionOption, 0, 3)
		for idx := 0; idx < 3; idx++ {
			optText := qData.Options[idx]
			opts = append(opts, domain.QuestionOption{
				ID:        fmt.Sprintf("opt_%d", idx+1),
				Text:      optText,
				IsCorrect: (idx == correctIdx),
			})
		}

		questions = append(questions, domain.Question{
			QuestionID:     cleanID,
			QuestionNumber: i + 1,
			TotalQuestions: len(concepts),
			ConceptID:      cleanID,
			ConceptName:    c.Name,
			Prompt:         qData.Prompt,
			Options:        opts,
		})
		correctAnswers[cleanID] = fmt.Sprintf("opt_%d", correctIdx+1)
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
		return nil, fmt.Errorf("create session: %w", err)
	}

	firstQuestion := &questions[0]
	return &StartResponse{
		SessionID:      session.SessionID,
		FirstQuestion:  firstQuestion,
		TotalQuestions: len(concepts),
	}, nil
}

// AnswerDiagnosticUseCase handles single answer submission and returns next state.
type AnswerDiagnosticUseCase struct {
	store SessionStore
}

func NewAnswerDiagnosticUseCase(store SessionStore) *AnswerDiagnosticUseCase {
	return &AnswerDiagnosticUseCase{store: store}
}

func (uc *AnswerDiagnosticUseCase) Execute(ctx context.Context, learnerID, sessionID, questionID, optionID string) (*domain.AnswerResponse, error) {
	questionID = strings.TrimSpace(questionID)
	optionID = strings.TrimSpace(optionID)
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
		if strings.TrimSpace(q.QuestionID) == questionID {
			currentQuestion = &q
			break
		}
	}
	if currentQuestion != nil {
		isValidOption := false
		for _, opt := range currentQuestion.Options {
			if strings.TrimSpace(opt.ID) == optionID {
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
	correctAnsID := session.CorrectAnswers[questionID]
	isCorrect := (optionID != "" && optionID == correctAnsID)

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
	return &domain.AnswerResponse{
		IsComplete:       session.Completed,
		IsCorrect:        isCorrect,
		CorrectOptionID:  correctAnsID,
		SelectedOptionID: optionID,
		NextQuestion:     next,
	}, nil
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

	prefs, err := uc.profile.GetPreferences(ctx, learnerID)
	priorLevel := "Beginner"
	if err == nil && prefs != nil && prefs.PriorExperience != "" {
		priorLevel = prefs.PriorExperience
	} else if p, pErr := uc.profile.GetPriorExperience(ctx, learnerID); pErr == nil && p != "" {
		priorLevel = p
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
				nodeID = strings.TrimSpace(c.NodeID)
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
	clean := strings.TrimSpace(nodeID)
	for _, c := range s.Concepts {
		if strings.TrimSpace(c.NodeID) == clean {
			return true
		}
	}
	return false
}

func indexOfConcept(s *domain.Session, nodeID string) int {
	clean := strings.TrimSpace(nodeID)
	for i, c := range s.Concepts {
		if strings.TrimSpace(c.NodeID) == clean {
			return i
		}
	}
	return -1
}

func buildFallbackQuestion(role, priorLevel, conceptName string) *QuestionData {
	cleanRole := strings.TrimSpace(strings.ToLower(role))
	switch strings.ToLower(priorLevel) {
	case "beginner":
		switch cleanRole {
		case "product_manager":
			return &QuestionData{
				Prompt: fmt.Sprintf("When starting fresh as a Product Manager, what is the primary focus of %s?", conceptName),
				Options: []string{
					"Understanding customer needs, defining product vision, and aligning business goals with user value.",
					"Writing low-level C++ memory allocation routines for hardware drivers.",
					"Configuring Kubernetes cluster ingress controllers and network routing.",
				},
				CorrectOption: 0,
			}
		case "mobile_engineer":
			return &QuestionData{
				Prompt: fmt.Sprintf("When starting fresh in Mobile Engineering, what is the core purpose of %s?", conceptName),
				Options: []string{
					"Understanding mobile OS lifecycles and choosing primary languages like Kotlin or Swift for app development.",
					"Managing physical database server racks in a data center.",
					"Writing server-side SQL stored procedures.",
				},
				CorrectOption: 0,
			}
		case "frontend_engineer":
			return &QuestionData{
				Prompt: fmt.Sprintf("As a beginner Frontend Engineer, what is the primary role of %s?", conceptName),
				Options: []string{
					"Building interactive web user interfaces using HTML structure, CSS styling, and JavaScript.",
					"Designing relational database schemas with primary keys.",
					"Configuring Linux kernel sysctl network parameters.",
				},
				CorrectOption: 0,
			}
		case "backend_engineer":
			return &QuestionData{
				Prompt: fmt.Sprintf("For a beginner Backend Engineer, what is the core purpose of %s?", conceptName),
				Options: []string{
					"Writing clean server-side logic, understanding data structures, and serving API requests.",
					"Creating CSS grid keyframe animations and responsive web layouts.",
					"Designing physical PCB circuit board layouts.",
				},
				CorrectOption: 0,
			}
		case "full_stack":
			return &QuestionData{
				Prompt: fmt.Sprintf("For a beginner Full Stack Developer, what is the main goal of %s?", conceptName),
				Options: []string{
					"Connecting client-side web frontends with server-side backends and database persistence.",
					"Assembling desktop computer hardware components.",
					"Writing GPU shader assembly code.",
				},
				CorrectOption: 0,
			}
		case "devops_sre":
			return &QuestionData{
				Prompt: fmt.Sprintf("As a beginner DevOps / SRE engineer, what is the primary focus of %s?", conceptName),
				Options: []string{
					"Automating software delivery, managing Linux environments, and configuring CI/CD pipelines.",
					"Designing vector graphic logos for marketing.",
					"Writing client-side React UI components.",
				},
				CorrectOption: 0,
			}
		case "ai_engineer":
			return &QuestionData{
				Prompt: fmt.Sprintf("For a beginner AI Engineer, what is the main purpose of %s?", conceptName),
				Options: []string{
					"Using Python libraries to load datasets, build foundational models, and integrate LLM APIs.",
					"Designing CSS flexbox page layouts.",
					"Managing physical network switches and cabling.",
				},
				CorrectOption: 0,
			}
		case "ai_data_scientist":
			return &QuestionData{
				Prompt: fmt.Sprintf("As a beginner AI Data Scientist, what is the primary goal of %s?", conceptName),
				Options: []string{
					"Analyzing datasets using Python, Pandas, and foundational statistical methods.",
					"Building Android app UI layouts with Jetpack Compose.",
					"Configuring Docker container bridge networks.",
				},
				CorrectOption: 0,
			}
		case "data_analyst":
			return &QuestionData{
				Prompt: fmt.Sprintf("For a beginner Data Analyst, what is the primary focus of %s?", conceptName),
				Options: []string{
					"Querying data with SQL, exploring metrics in Excel, and building visual reporting dashboards.",
					"Writing low-level C memory management routines.",
					"Building iOS apps with Swift.",
				},
				CorrectOption: 0,
			}
		case "machine_learning":
			return &QuestionData{
				Prompt: fmt.Sprintf("As a beginner Machine Learning engineer, what is the core focus of %s?", conceptName),
				Options: []string{
					"Using Python and basic math/statistics to collect, preprocess, and analyze training data.",
					"Designing HTML/CSS marketing web pages.",
					"Configuring Linux network firewalls.",
				},
				CorrectOption: 0,
			}
		default:
			return &QuestionData{
				Prompt: fmt.Sprintf("When starting fresh in %s, what is the primary foundational goal of %s?", strings.ReplaceAll(cleanRole, "_", " "), conceptName),
				Options: []string{
					fmt.Sprintf("Understanding basic component roles, core definitions, and entry-level principles of %s.", conceptName),
					"Replacing physical motherboard BIOS hardware chips.",
					"Writing client-side CSS keyframe animations.",
				},
				CorrectOption: 0,
			}
		}
	case "advanced":
		return &QuestionData{
			Prompt: fmt.Sprintf("In high-scale %s architectures, which statement best characterizes the concurrency, performance trade-offs, and internal mechanics of %s?", strings.ReplaceAll(cleanRole, "_", " "), conceptName),
			Options: []string{
				fmt.Sprintf("It governs critical architectural constraints, state synchronization, and execution efficiency for %s.", conceptName),
				"It eliminates memory allocation overhead by running directly on GPU registers.",
				"It acts solely as a synchronous CSV export utility.",
			},
			CorrectOption: 0,
		}
	default:
		return &QuestionData{
			Prompt: fmt.Sprintf("In standard %s projects, how is %s typically applied in modern developer workflows?", strings.ReplaceAll(cleanRole, "_", " "), conceptName),
			Options: []string{
				fmt.Sprintf("It structures practical application components and standard design patterns for %s.", conceptName),
				"It bypasses network security protocols completely.",
				"It is used exclusively for formatting static CSS text styles.",
			},
			CorrectOption: 0,
		}
	}
}
