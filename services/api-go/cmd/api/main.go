package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	"github.com/hcl-backend/services/api-go/internal/aiclient"
	"github.com/hcl-backend/services/api-go/internal/aiengine"
	"github.com/hcl-backend/services/api-go/internal/aihttp"
	"github.com/hcl-backend/services/api-go/internal/bootstrap"
	"github.com/hcl-backend/services/api-go/internal/platform/auditlog"
	"github.com/hcl-backend/services/api-go/internal/platform/cache"
	"github.com/hcl-backend/services/api-go/internal/platform/config"
	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/hcl-backend/services/api-go/internal/platform/events"
	"github.com/hcl-backend/services/api-go/internal/platform/httpserver"
	"github.com/hcl-backend/services/api-go/internal/platform/logger"
	"github.com/hcl-backend/services/api-go/internal/platform/middleware"

	adminApp "github.com/hcl-backend/services/api-go/internal/admin/application"
	adminDomain "github.com/hcl-backend/services/api-go/internal/admin/domain"
	adminInfra "github.com/hcl-backend/services/api-go/internal/admin/infrastructure"
	adminHttp "github.com/hcl-backend/services/api-go/internal/admin/interfaces/http"

	assessmentApp "github.com/hcl-backend/services/api-go/internal/assessment/application"
	assessmentDomain "github.com/hcl-backend/services/api-go/internal/assessment/domain"
	assessmentInfra "github.com/hcl-backend/services/api-go/internal/assessment/infrastructure"
	assessmentHttp "github.com/hcl-backend/services/api-go/internal/assessment/interfaces/http"

	competencyApp "github.com/hcl-backend/services/api-go/internal/competency/application"
	competencyInfra "github.com/hcl-backend/services/api-go/internal/competency/infrastructure"
	competencyHttp "github.com/hcl-backend/services/api-go/internal/competency/interfaces/http"

	diagApp "github.com/hcl-backend/services/api-go/internal/diagnostics/application"
	diagInfra "github.com/hcl-backend/services/api-go/internal/diagnostics/infrastructure"
	diagHttp "github.com/hcl-backend/services/api-go/internal/diagnostics/interfaces/http"

	feedbackApp "github.com/hcl-backend/services/api-go/internal/feedback/application"
	feedbackInfra "github.com/hcl-backend/services/api-go/internal/feedback/infrastructure"
	feedbackHttp "github.com/hcl-backend/services/api-go/internal/feedback/interfaces/http"

	goalsApp "github.com/hcl-backend/services/api-go/internal/goals/application"
	goalsInfra "github.com/hcl-backend/services/api-go/internal/goals/infrastructure"
	goalsHttp "github.com/hcl-backend/services/api-go/internal/goals/interfaces/http"

	identityApp "github.com/hcl-backend/services/api-go/internal/identity/application"
	identityInfra "github.com/hcl-backend/services/api-go/internal/identity/infrastructure"
	identityHttp "github.com/hcl-backend/services/api-go/internal/identity/interfaces/http"

	knowledgeApp "github.com/hcl-backend/services/api-go/internal/knowledge/application"
	knowledgeInfra "github.com/hcl-backend/services/api-go/internal/knowledge/infrastructure"
	knowledgeHttp "github.com/hcl-backend/services/api-go/internal/knowledge/interfaces/http"

	learnerApp "github.com/hcl-backend/services/api-go/internal/learner/application"
	learnerInfra "github.com/hcl-backend/services/api-go/internal/learner/infrastructure"
	learnerHttp "github.com/hcl-backend/services/api-go/internal/learner/interfaces/http"

	notifApp "github.com/hcl-backend/services/api-go/internal/notifications/application"
	notifInfra "github.com/hcl-backend/services/api-go/internal/notifications/infrastructure"
	notifHttp "github.com/hcl-backend/services/api-go/internal/notifications/interfaces/http"

	progressApp "github.com/hcl-backend/services/api-go/internal/progress/application"
	progressDomain "github.com/hcl-backend/services/api-go/internal/progress/domain"
	progressInfra "github.com/hcl-backend/services/api-go/internal/progress/infrastructure"
	progressHttp "github.com/hcl-backend/services/api-go/internal/progress/interfaces/http"

	projectsApp "github.com/hcl-backend/services/api-go/internal/projects/application"
	projectsDomain "github.com/hcl-backend/services/api-go/internal/projects/domain"
	projectsHttp "github.com/hcl-backend/services/api-go/internal/projects/interfaces/http"

	resourceApp "github.com/hcl-backend/services/api-go/internal/resources/application"
	resourceInfra "github.com/hcl-backend/services/api-go/internal/resources/infrastructure"
	resourceHttp "github.com/hcl-backend/services/api-go/internal/resources/interfaces/http"

	roadmapApp "github.com/hcl-backend/services/api-go/internal/roadmap/application"
	roadmapInfra "github.com/hcl-backend/services/api-go/internal/roadmap/infrastructure"
	roadmapHttp "github.com/hcl-backend/services/api-go/internal/roadmap/interfaces/http"
)

func main() {
	seed := flag.Bool("seed", false, "seed the platform baseline data (users, domains, concepts, resources) before serving")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.App.LogLevel)
	log.Info("Starting API service", "env", cfg.App.Env)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize Database
	dbPool, err := database.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer dbPool.Close()

	// Initialize Redis Cache
	redisClient, err := cache.NewClient(ctx, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal("Failed to connect to redis", "error", err)
	}
	defer redisClient.Close()
	log.Info("Connected to PostgreSQL and Redis")

	// Initialize the ported FastAPI AI intelligence engine (graph + BKT + LLM +
	// guardrails + sentiment + voice + RAG). Fatal on startup so a malformed
	// knowledge base is surfaced immediately, like FastAPI's import-time load.
	aiApp, err := aiengine.GetApp()
	if err != nil {
		log.Fatal("Failed to initialize AI intelligence engine", "error", err)
	}
	log.Info("AI intelligence engine initialized",
		"domains", len(aiApp.Graph.DomainList),
		"llm_provider", aiApp.LLM.Provider())

	if *seed {
		seedCtx, seedCancel := context.WithTimeout(context.Background(), 120*time.Second)
		if err := bootstrap.Seed(seedCtx, dbPool, aiApp.Graph, log.Slog()); err != nil {
			log.Fatal("Failed to seed platform baseline data", "error", err)
		}
		seedCancel()
		log.Info("Platform baseline data seeded")
	}

	audit := auditlog.NewPostgresAuditLog(dbPool)
	aiClient := aiclient.NewLocalClient(aiApp, dbPool)

	// Setup Identity Module
	userRepo := identityInfra.NewPostgresUserRepository(dbPool)
	authSecret := cfg.App.Env // Using env as mock secret for now
	authService := identityApp.NewAuthService(userRepo, authSecret, audit)
	identityHandler := identityHttp.NewHandler(authService)

	// Setup Learner Module
	learnerRepo := learnerInfra.NewPostgresLearnerProfileRepository(dbPool)
	updatePreferencesUseCase := learnerApp.NewUpdatePreferencesUseCase(learnerRepo)
	updateSettingsUseCase := learnerApp.NewUpdateSettingsUseCase(learnerInfra.NewPostgresSettingsRepository(dbPool))
	learnerHandler := learnerHttp.NewHandler(updatePreferencesUseCase, updateSettingsUseCase)

	// Setup Knowledge Module
	knowledgeRepo := knowledgeInfra.NewPostgresKnowledgeRepository(dbPool)
	knowledgeMeta := knowledgeInfra.NewGraphMetaProvider(aiApp.Graph)
	knowledgeRAG := &knowledgeRAGService{engine: aiApp.Graph}
	knowledgeService := knowledgeApp.NewKnowledgeService(knowledgeRepo, knowledgeMeta, knowledgeRAG)
	knowledgeHandler := knowledgeHttp.NewHandler(knowledgeService)

	// Setup Feedback Module
	feedbackRepo := feedbackInfra.NewPostgresFeedbackRepository(dbPool)
	recordFeedbackUseCase := feedbackApp.NewRecordResourceFeedbackUseCase(feedbackRepo)
	feedbackHandler := feedbackHttp.NewHandler(recordFeedbackUseCase)

	// Setup Goals Module
	goalRepo := goalsInfra.NewPostgresGoalRepository(dbPool)
	goalsKnowledge := &goalsKnowledgeAdapter{svc: knowledgeService}
	createGoalUseCase := goalsApp.NewCreateGoalUseCase(goalRepo, aiClient, goalsKnowledge)
	getCurrentGoalUseCase := goalsApp.NewGetCurrentGoalUseCase(goalRepo, knowledgeService)
	goalsHandler := goalsHttp.NewHandler(createGoalUseCase, getCurrentGoalUseCase)

	// Setup Competency Module
	competencyRepo := competencyInfra.NewPostgresCompetencyRepository(dbPool)
	updateCompetencyUseCase := competencyApp.NewUpdateCompetencyUseCase(competencyRepo)
	getCompetencyDetailUseCase := competencyApp.NewGetCompetencyDetailUseCase(competencyRepo, knowledgeService)
	getCompetencyHistoryUseCase := competencyApp.NewGetCompetencyHistoryUseCase(competencyRepo, knowledgeService)
	competencyHandler := competencyHttp.NewHandler(getCompetencyDetailUseCase, getCompetencyHistoryUseCase)

	// Setup Progress Module
	progressRepo := progressInfra.NewPostgresProgressRepository(dbPool)
	txManager := database.NewTxManager(dbPool)
	recordEvidenceUseCase := progressApp.NewRecordEvidenceUseCase(txManager, progressRepo, updateCompetencyUseCase)
	recordEngagementUseCase := progressApp.NewRecordEngagementUseCase(progressRepo)
	progressGoals := &progressGoalsService{repo: goalRepo}
	getSummaryUseCase := progressApp.NewGetProgressSummaryUseCase(progressRepo, progressGoals)
	getGoalSummaryUseCase := progressApp.NewGetGoalCompletionSummaryUseCase(progressRepo, progressGoals)
	progressHandler := progressHttp.NewHandler(recordEvidenceUseCase, recordEngagementUseCase, getSummaryUseCase, getGoalSummaryUseCase)

	// Setup Resources Module
	resourcesRepo := resourceInfra.NewPostgresResourceRepository(dbPool)
	resourceValidator := &resourceConceptValidator{svc: knowledgeService}
	createResourceUseCase := resourceApp.NewCreateResourceUseCase(resourcesRepo, resourceValidator)
	updateResourceUseCase := resourceApp.NewUpdateResourceUseCase(resourcesRepo)
	listResourcesUseCase := resourceApp.NewListResourcesUseCase(resourcesRepo)
	getFeedbackSignalsUseCase := resourceApp.NewGetFeedbackSignalsUseCase(resourcesRepo)
	getAlternateUseCase := resourceApp.NewGetAlternateResourcesUseCase(resourcesRepo, aiClient)
	explainUseCase := resourceApp.NewExplainResourceRelevanceUseCase(aiClient)
	resourcesHandler := resourceHttp.NewHandler(createResourceUseCase, updateResourceUseCase, listResourcesUseCase, getFeedbackSignalsUseCase, getAlternateUseCase, explainUseCase)

	// Setup Assessment Module (deterministic quiz generation + the real
	// evidence pipeline for competency transitions).
	assessmentRepo := assessmentInfra.NewPostgresAssessmentRepository(dbPool)
	getAssessmentUseCase := assessmentApp.NewGetAssessmentUseCase(assessmentRepo, knowledgeService)
	assessmentAI := &assessmentAIClient{client: aiClient}
	assessmentEvidence := &assessmentEvidenceService{uc: recordEvidenceUseCase}
	submitAssessmentUseCase := assessmentApp.NewSubmitAssessmentUseCase(assessmentRepo, knowledgeService, assessmentAI, assessmentEvidence)
	assessmentHandler := assessmentHttp.NewHandler(getAssessmentUseCase, submitAssessmentUseCase)

	// Setup Roadmap Module
	roadmapRepo := roadmapInfra.NewPostgresRoadmapRepository(dbPool)
	roadmapGoalsSvc := &roadmapGoalsService{repo: goalRepo}
	roadmapCompetencySvc := &roadmapCompetencyService{repo: competencyRepo}
	roadmapAILocal := &roadmapApp.LocalAIClientService{App: aiApp, DbPool: dbPool}
	regenerateRoadmapUseCase := roadmapApp.NewRegenerateRoadmapUseCase(
		txManager, roadmapRepo, roadmapGoalsSvc, roadmapCompetencySvc, roadmapAILocal,
	)
	getActiveRoadmapUseCase := roadmapApp.NewGetActiveRoadmapUseCase(roadmapRepo)
	getConceptExplanationUseCase := roadmapApp.NewGetConceptExplanationUseCase(roadmapAILocal)
	getDailyTasksUseCase := roadmapApp.NewGetDailyTasksUseCase(roadmapRepo)
	toggleDailyTaskUseCase := roadmapApp.NewToggleDailyTaskUseCase(roadmapRepo)
	roadmapHandler := roadmapHttp.NewHandler(getActiveRoadmapUseCase, regenerateRoadmapUseCase, getConceptExplanationUseCase, getDailyTasksUseCase, toggleDailyTaskUseCase)

	// Setup Diagnostics Module (short-lived in-memory onboarding sessions).
	diagStore := diagInfra.NewInMemorySessionStore()
	diagGoals := &diagGoalsService{repo: goalRepo}
	diagResolver := &diagResolver{svc: knowledgeService}
	diagGraph := &diagGraphService{engine: aiApp.Graph}
	diagProfile := &diagProfileService{db: dbPool}
	diagLLM := &diagLLMService{llm: aiApp.LLM}
	diagCompetency := &diagCompetencyService{repo: competencyRepo}
	startDiagnosticUseCase := diagApp.NewStartDiagnosticUseCase(diagStore, diagGoals, diagResolver, diagGraph, diagProfile, diagLLM)
	answerDiagnosticUseCase := diagApp.NewAnswerDiagnosticUseCase(diagStore)
	resultsDiagnosticUseCase := diagApp.NewResultsDiagnosticUseCase(diagStore, diagGoals, diagResolver, diagProfile, diagGraph, diagLLM, diagCompetency)
	diagnosticsHandler := diagHttp.NewHandler(startDiagnosticUseCase, answerDiagnosticUseCase, resultsDiagnosticUseCase)

	// Setup Projects Module (storage/artifact wiring remains mocked).
	projectsAssessmentMock := &mockProjectsAssessmentService{}
	projectsEvidenceMock := &mockProjectsEvidenceService{}
	projectsStorageMock := &mockProjectsStorageService{}
	getProjectUseCase := projectsApp.NewGetProjectUseCase(projectsAssessmentMock)
	submitProjectUseCase := projectsApp.NewSubmitProjectUseCase(projectsAssessmentMock, projectsStorageMock, projectsEvidenceMock)
	getProjectStatusUseCase := projectsApp.NewGetProjectStatusUseCase(projectsEvidenceMock)
	projectsHandler := projectsHttp.NewHandler(getProjectUseCase, submitProjectUseCase, getProjectStatusUseCase)

	// Setup Notifications Module
	notificationsRepo := notifInfra.NewPostgresRepository(dbPool)
	getNotificationsUseCase := notifApp.NewGetNotificationsUseCase(notificationsRepo)
	markNotificationReadUseCase := notifApp.NewMarkNotificationReadUseCase(notificationsRepo)
	notificationsHandler := notifHttp.NewHandler(getNotificationsUseCase, markNotificationReadUseCase)

	// Setup Event Bus
	redisBus := events.NewRedisBus(redisClient)
	_ = redisBus // To avoid unused variable error if not injected yet.

	// Setup Admin Module
	adminIdentityMock := &mockAdminIdentityService{}
	auditRepo := adminInfra.NewPostgresAuditRepository(dbPool)
	listUsersUseCase := adminApp.NewListUsersUseCase(adminIdentityMock)
	updateUserUseCase := adminApp.NewUpdateUserUseCase(adminIdentityMock, auditRepo)
	getAuditLogUseCase := adminApp.NewGetAuditLogUseCase(auditRepo)
	adminHandler := adminHttp.NewHandler(listUsersUseCase, updateUserUseCase, getAuditLogUseCase)

	// Setup Router
	mux := http.NewServeMux()
	identityHandler.RegisterRoutes(mux)
	learnerHandler.RegisterRoutes(mux)
	knowledgeHandler.RegisterRoutes(mux)
	feedbackHandler.RegisterRoutes(mux)
	goalsHandler.RegisterRoutes(mux)
	assessmentHandler.RegisterRoutes(mux)
	competencyHandler.RegisterRoutes(mux)
	progressHandler.RegisterRoutes(mux)
	resourcesHandler.RegisterRoutes(mux)
	roadmapHandler.RegisterRoutes(mux)
	projectsHandler.RegisterRoutes(mux)
	diagnosticsHandler.RegisterRoutes(mux)
	notificationsHandler.RegisterRoutes(mux)
	adminHandler.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Ported FastAPI intelligence service routers (exact /api/v1 paths).
	aiHandler := &aihttp.Handler{App: aiApp}
	aiHandler.RegisterRoutes(mux)

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := dbPool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unavailable", "database": "disconnected"})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	// Setup Middleware chain
	handler := middleware.RequestID(mux)
	handler = middleware.Logger(log)(handler)
	handler = middleware.Recovery(handler)
	handler = middleware.Auth(authSecret)(handler) // Use the same JWT secret

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // In prod, this should be restricted
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
	}).Handler(handler)

	// Start Server
	server := httpserver.New(cfg.App.Port, corsHandler, log)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Server error", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server stopped")
}

// goalsKnowledgeAdapter adapts the knowledge module's ResolveStructure to the
// goals module port (distinct application-layer types).
type goalsKnowledgeAdapter struct {
	svc *knowledgeApp.KnowledgeService
}

func (a *goalsKnowledgeAdapter) ResolveStructure(ctx context.Context, ref string) (*goalsApp.ResolvedStructure, error) {
	rs, err := a.svc.ResolveStructure(ctx, ref)
	if err != nil {
		return nil, err
	}
	return &goalsApp.ResolvedStructure{
		ID:          rs.ID,
		DomainSlug:  rs.DomainSlug,
		DomainName:  rs.DomainName,
		Confidence:  rs.Confidence,
		IsPublished: rs.IsPublished,
	}, nil
}

// progressGoalsService adapts the goals repository to the progress module's
// GoalService port (scopes the progress dashboard to the active goal).
type progressGoalsService struct {
	repo *goalsInfra.PostgresGoalRepository
}

func (s *progressGoalsService) ActiveStructureMeta(ctx context.Context, learnerID string) (string, string, string, error) {
	g, err := s.repo.GetActiveByLearnerID(ctx, learnerID)
	if err != nil {
		return "", "", "", err
	}
	return g.ID, g.GoalText, g.KnowledgeStructureID, nil
}

// roadmapGoalsService adapts the REAL goals repository to the Roadmap module's
// GoalsService port (drives roadmap regeneration from the learner's active goal).
type roadmapGoalsService struct {
	repo *goalsInfra.PostgresGoalRepository
}

func (s *roadmapGoalsService) GetActiveGoal(ctx context.Context, learnerID string) (roadmapApp.Goal, error) {
	g, err := s.repo.GetActiveByLearnerID(ctx, learnerID)
	if err != nil {
		return roadmapApp.Goal{}, err
	}
	return roadmapApp.Goal{
		ID:                   g.ID,
		KnowledgeStructureID: g.KnowledgeStructureID,
		GoalText:             g.GoalText,
	}, nil
}

// roadmapCompetencyService adapts the competency repository (node id -> state)
// to the roadmap proposal generator.
type roadmapCompetencyService struct {
	repo *competencyInfra.PostgresCompetencyRepository
}

func (s *roadmapCompetencyService) GetLearnerCompetencies(ctx context.Context, learnerID string) (map[string]string, error) {
	records, err := s.repo.ListByLearner(ctx, learnerID)
	if err != nil {
		return nil, err
	}
	stateByNode := make(map[string]string, len(records))
	for _, r := range records {
		stateByNode[r.ConceptID] = r.State
	}
	return stateByNode, nil
}

// resourceConceptValidator adapts the knowledge module's concept validation to
// the resources module port.
type resourceConceptValidator struct {
	svc *knowledgeApp.KnowledgeService
}

func (v *resourceConceptValidator) ValidateConcepts(ctx context.Context, conceptIDs []string) error {
	for _, id := range conceptIDs {
		if err := v.svc.ValidateConcept(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// assessmentAIClient adapts the aiengine local evaluation client to the
// assessment module port (free-text grading against the node's rubric).
type assessmentAIClient struct {
	client *aiclient.LocalClient
}

func (a *assessmentAIClient) Evaluate(ctx context.Context, conceptID, domainID string, submission json.RawMessage) (*assessmentDomain.EvaluationResult, error) {
	var sub struct {
		FreeText string `json:"freeText"`
	}
	if err := json.Unmarshal(submission, &sub); err != nil || sub.FreeText == "" {
		return nil, fmt.Errorf("free-text submission requires a non-empty answer")
	}
	return a.client.Evaluate(ctx, mustJSON(map[string]any{
		"domain_id":      domainID,
		"node_id":        conceptID,
		"student_answer": sub.FreeText,
	}), nil)
}

// assessmentEvidenceService routes assessment evidence through the progress
// pipeline so competency state and roadmap nodes update atomically.
type assessmentEvidenceService struct {
	uc *progressApp.RecordEvidenceUseCase
}

func (s *assessmentEvidenceService) RecordEvidence(ctx context.Context, evidence *assessmentDomain.Evidence) (string, error) {
	return s.uc.RecordEvidence(ctx, &progressDomain.Evidence{
		LearnerID:              evidence.LearnerID,
		ConceptID:              evidence.ConceptID,
		AssessmentDefinitionID: evidence.AssessmentDefinitionID,
		SubmissionData:         evidence.SubmissionData,
		Score:                  evidence.Score,
		Confidence:             evidence.Confidence,
		EvaluatorType:          evidence.EvaluatorType,
		Result:                 evidence.Result,
	})
}

// diagGoalsService adapts the goals repository for diagnostics (target the
// active goal's structure).
type diagGoalsService struct {
	repo *goalsInfra.PostgresGoalRepository
}

func (s *diagGoalsService) ActiveStructure(ctx context.Context, learnerID string) (string, string, error) {
	g, err := s.repo.GetActiveByLearnerID(ctx, learnerID)
	if err != nil {
		return "", "", err
	}
	return g.GoalText, g.KnowledgeStructureID, nil
}

// diagResolver maps a structure uuid to its roadmap.sh domain slug.
type diagResolver struct {
	svc *knowledgeApp.KnowledgeService
}

func (r *diagResolver) StructureDomainSlug(ctx context.Context, structureID string) (string, error) {
	slug, _, err := r.svc.StructureName(ctx, structureID)
	return slug, err
}

// diagGraphService reads the deterministic topological concept order.
type diagGraphService struct {
	engine *aiengine.GraphEngine
}

func (g *diagGraphService) TopoConcepts(ctx context.Context, domainSlug string) ([]diagApp.NodeRef, error) {
	dg, ok := g.engine.DomainGraphs[domainSlug]
	if !ok {
		return nil, fmt.Errorf("unknown domain: %s", domainSlug)
	}
	order := dg.TopoOrder()
	refs := make([]diagApp.NodeRef, 0, len(order))
	for _, id := range order {
		node, ok := dg.NodeByID[id]
		if !ok {
			continue
		}
		refs = append(refs, diagApp.NodeRef{NodeID: id, Name: node.Name})
	}
	return refs, nil
}

func (g *diagGraphService) GetResources(domainSlug, nodeID string) []string {
	resources := g.engine.GetResourcesForNode(domainSlug, nodeID, 2)
	var textChunks []string
	for _, r := range resources {
		if txt, ok := r.Metadata["document_text"].(string); ok && txt != "" {
			textChunks = append(textChunks, txt)
		}
	}
	return textChunks
}

type diagProfileService struct {
	db *pgxpool.Pool
}

func (s *diagProfileService) GetPriorExperience(ctx context.Context, learnerID string) (string, error) {
	var priorExperience string
	err := s.db.QueryRow(ctx, "SELECT prior_experience FROM platform.learner_profiles WHERE user_id = $1", learnerID).Scan(&priorExperience)
	if err != nil {
		return "Beginner", nil
	}
	return priorExperience, nil
}

func (s *diagProfileService) GetRole(ctx context.Context, learnerID string) (string, error) {
	var role string
	err := s.db.QueryRow(ctx, "SELECT role FROM platform.learner_profiles WHERE user_id = $1", learnerID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

type diagCompetencyService struct {
	repo *competencyInfra.PostgresCompetencyRepository
}

func (s *diagCompetencyService) SaveBaseline(ctx context.Context, learnerID, nodeID, state string) error {
	return s.repo.CreateBaseline(ctx, learnerID, nodeID, state)
}

type knowledgeRAGService struct {
	engine *aiengine.GraphEngine
}

func (s *knowledgeRAGService) GetRAGContext(domainID, nodeID string) []string {
	res := s.engine.GetResourcesForNode(domainID, nodeID, 2)
	var chunks []string
	for _, r := range res {
		if docText, ok := r.Metadata["document_text"].(string); ok && docText != "" {
			chunks = append(chunks, docText)
		}
	}
	return chunks
}

type diagLLMService struct {
	llm *aiengine.LLMClient
}

func (s *diagLLMService) GenerateQuestionPrompt(ctx context.Context, role, priorLevel, conceptName, ragContext string) (*diagApp.QuestionData, error) {
	systemPrompt := "You are a professional technical interviewer assessing a candidate."
	level := strings.ToLower(priorLevel)
	var userPrompt string

	if level == "beginner" {
		userPrompt = fmt.Sprintf(`Generate a single basic yes/no concept knowledge question to gauge if the candidate knows the concept: %q.
Role: %s
Authoritative Reference Text (RAG Context):
%s

Instructions:
1. Formulate a basic question asking if they know the concept (e.g. "Do you know how to declare a list in Python?" or similar basic question).
2. The options must be exactly ["Yes", "No"].
3. Set correct_option to 0 (which corresponds to "Yes", indicating they know it).`, conceptName, role, ragContext)
	} else if level == "intermediate" {
		userPrompt = fmt.Sprintf(`Generate a single syntax or code logic multiple choice question to gauge the candidate's understanding of the concept: %q.
Role: %s
Authoritative Reference Text (RAG Context):
%s

Instructions:
1. Formulate a concrete syntax or logic question with 5 options.
2. The question must be related to the role and concept based on the reference text.
3. Provide exactly 5 options. One must be correct, the other four must be plausible but incorrect options.
4. Set correct_option to the 0-based index of the correct option (0 to 4).`, conceptName, role, ragContext)
	} else {
		userPrompt = fmt.Sprintf(`Generate a single advanced concept multiple choice question to gauge the candidate's understanding of the concept: %q.
Role: %s
Authoritative Reference Text (RAG Context):
%s

Instructions:
1. Formulate an advanced, deep concept question with 5 options.
2. The question must be strictly based on the reference text and highly specific to advanced scenarios of the role.
3. Provide exactly 5 options. One must be correct, the other four must be plausible but incorrect options.
4. Set correct_option to the 0-based index of the correct option (0 to 4).`, conceptName, role, ragContext)
	}

	responseSchema := map[string]any{
		"prompt": map[string]any{
			"type":        "string",
			"description": "The question prompt text.",
		},
		"options": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "string",
			},
			"description": "The list of option choices.",
		},
		"correct_option": map[string]any{
			"type":        "integer",
			"description": "The 0-based index of the correct option.",
		},
	}

	result := s.llm.GenerateStructuredJSON(systemPrompt, userPrompt, responseSchema)
	prompt, _ := result["prompt"].(string)
	if prompt == "" {
		if errStr, ok := result["error"].(string); ok {
			return nil, fmt.Errorf("LLM error: %s", errStr)
		}
		return nil, fmt.Errorf("invalid LLM response format: prompt is empty")
	}

	var options []string
	if optsRaw, ok := result["options"].([]any); ok {
		for _, o := range optsRaw {
			if sVal, ok := o.(string); ok {
				options = append(options, sVal)
			}
		}
	} else if optsRawStr, ok := result["options"].([]string); ok {
		options = optsRawStr
	}

	if len(options) == 0 {
		if level == "beginner" {
			options = []string{"Yes", "No"}
		} else {
			options = []string{"Option A", "Option B", "Option C", "Option D", "Option E"}
		}
	}

	correctIdx := 0
	if fIdx, ok := result["correct_option"].(float64); ok {
		correctIdx = int(fIdx)
	} else if iIdx, ok := result["correct_option"].(int); ok {
		correctIdx = iIdx
	}

	if correctIdx < 0 || correctIdx >= len(options) {
		correctIdx = 0
	}

	return &diagApp.QuestionData{
		Prompt:        prompt,
		Options:       options,
		CorrectOption: correctIdx,
	}, nil
}

func (s *diagLLMService) GenerateWeakAreasExplanation(ctx context.Context, role, priorLevel string, gaps []string, ragContext string) (string, error) {
	systemPrompt := "You are a friendly, encouraging learning advisor."
	userPrompt := fmt.Sprintf(`Analyze the candidate's diagnostic results and explain their weak areas.
Role: %s
Experience Level: %s
Identified Gaps (Weak Concepts): %s
Authoritative Reference Text (RAG Context):
%s

Instructions:
1. Explain why these concepts are crucial for the %s role.
2. Outline what they should focus on first to bridge these gaps, referencing the RAG content context.
3. Be concise and write in a clear, supportive tone.
4. Keep the explanation within 3 paragraphs.
5. Output ONLY the explanation text as plain text. Do not wrap in markdown or JSON.`,
		role, priorLevel, strings.Join(gaps, ", "), ragContext, role)

	responseSchema := map[string]any{
		"explanation": map[string]any{
			"type": "string",
			"description": "The explanation of the candidate's weak areas and baseline results.",
		},
	}

	result := s.llm.GenerateStructuredJSON(systemPrompt, userPrompt, responseSchema)
	if explanation, ok := result["explanation"].(string); ok && explanation != "" {
		return explanation, nil
	}
	if errStr, ok := result["error"].(string); ok {
		return "", fmt.Errorf("LLM error: %s", errStr)
	}
	return "", fmt.Errorf("invalid LLM response format")
}

func mustJSON(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return raw
}

type mockProjectsAssessmentService struct{}

func (m *mockProjectsAssessmentService) GetProjectDefinition(ctx context.Context, conceptID string) (*projectsDomain.Project, error) {
	return &projectsDomain.Project{ID: "proj-1", ConceptID: conceptID}, nil
}

type mockProjectsEvidenceService struct{}

func (m *mockProjectsEvidenceService) RecordProjectSubmission(ctx context.Context, submission *projectsDomain.ProjectSubmission) error {
	return nil
}
func (m *mockProjectsEvidenceService) GetProjectStatus(ctx context.Context, learnerID, conceptID string) (*projectsDomain.ProjectState, error) {
	return &projectsDomain.ProjectState{Status: projectsDomain.ProjectStatusPending}, nil
}

type mockProjectsStorageService struct{}

func (m *mockProjectsStorageService) ValidateArtifactReference(ctx context.Context, reference string) error {
	return nil
}

type mockAdminIdentityService struct{}

func (m *mockAdminIdentityService) ListUsers(ctx context.Context) ([]adminDomain.User, error) {
	return []adminDomain.User{{ID: "mock-user-1", Role: "learner"}}, nil
}
func (m *mockAdminIdentityService) GetUserByID(ctx context.Context, id string) (*adminDomain.User, error) {
	return &adminDomain.User{ID: id, Role: "learner"}, nil
}
func (m *mockAdminIdentityService) UpdateUser(ctx context.Context, id string, role, status *string) (*adminDomain.User, error) {
	return &adminDomain.User{ID: id, Role: "curator"}, nil
}
