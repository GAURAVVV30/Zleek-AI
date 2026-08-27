package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hcl-backend/services/api-go/internal/platform/cache"
	"github.com/hcl-backend/services/api-go/internal/platform/config"
	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/hcl-backend/services/api-go/internal/platform/httpserver"
	"github.com/hcl-backend/services/api-go/internal/platform/logger"
	"github.com/hcl-backend/services/api-go/internal/platform/middleware"
	"github.com/rs/cors"

	"github.com/hcl-backend/services/api-go/internal/aiclient"
	assessmentApp "github.com/hcl-backend/services/api-go/internal/assessment/application"
	assessmentInfra "github.com/hcl-backend/services/api-go/internal/assessment/infrastructure"
	assessmentHttp "github.com/hcl-backend/services/api-go/internal/assessment/interfaces/http"
	"github.com/hcl-backend/services/api-go/internal/goals/application"
	"github.com/hcl-backend/services/api-go/internal/goals/infrastructure"
	goalsHttp "github.com/hcl-backend/services/api-go/internal/goals/interfaces/http"

	competencyApp "github.com/hcl-backend/services/api-go/internal/competency/application"
	competencyInfra "github.com/hcl-backend/services/api-go/internal/competency/infrastructure"
	competencyHttp "github.com/hcl-backend/services/api-go/internal/competency/interfaces/http"

	progressApp "github.com/hcl-backend/services/api-go/internal/progress/application"
	progressInfra "github.com/hcl-backend/services/api-go/internal/progress/infrastructure"
	progressHttp "github.com/hcl-backend/services/api-go/internal/progress/interfaces/http"

	resourcesApp "github.com/hcl-backend/services/api-go/internal/resources/application"
	resourcesInfra "github.com/hcl-backend/services/api-go/internal/resources/infrastructure"
	resourcesHttp "github.com/hcl-backend/services/api-go/internal/resources/interfaces/http"

	roadmapApp "github.com/hcl-backend/services/api-go/internal/roadmap/application"
	roadmapInfra "github.com/hcl-backend/services/api-go/internal/roadmap/infrastructure"
	roadmapHttp "github.com/hcl-backend/services/api-go/internal/roadmap/interfaces/http"

	projectsApp "github.com/hcl-backend/services/api-go/internal/projects/application"
	projectsDomain "github.com/hcl-backend/services/api-go/internal/projects/domain"
	projectsHttp "github.com/hcl-backend/services/api-go/internal/projects/interfaces/http"

	notifApp "github.com/hcl-backend/services/api-go/internal/notifications/application"
	notifInfra "github.com/hcl-backend/services/api-go/internal/notifications/infrastructure"
	notifHttp "github.com/hcl-backend/services/api-go/internal/notifications/interfaces/http"

	adminApp "github.com/hcl-backend/services/api-go/internal/admin/application"
	adminDomain "github.com/hcl-backend/services/api-go/internal/admin/domain"
	adminInfra "github.com/hcl-backend/services/api-go/internal/admin/infrastructure"
	adminHttp "github.com/hcl-backend/services/api-go/internal/admin/interfaces/http"

	identityApp "github.com/hcl-backend/services/api-go/internal/identity/application"
	identityInfra "github.com/hcl-backend/services/api-go/internal/identity/infrastructure"
	identityHttp "github.com/hcl-backend/services/api-go/internal/identity/interfaces/http"

	learnerApp "github.com/hcl-backend/services/api-go/internal/learner/application"
	learnerInfra "github.com/hcl-backend/services/api-go/internal/learner/infrastructure"
	learnerHttp "github.com/hcl-backend/services/api-go/internal/learner/interfaces/http"

	knowledgeApp "github.com/hcl-backend/services/api-go/internal/knowledge/application"
	knowledgeInfra "github.com/hcl-backend/services/api-go/internal/knowledge/infrastructure"
	knowledgeHttp "github.com/hcl-backend/services/api-go/internal/knowledge/interfaces/http"

	feedbackApp "github.com/hcl-backend/services/api-go/internal/feedback/application"
	feedbackInfra "github.com/hcl-backend/services/api-go/internal/feedback/infrastructure"
	feedbackHttp "github.com/hcl-backend/services/api-go/internal/feedback/interfaces/http"

	"github.com/hcl-backend/services/api-go/internal/platform/events"
)

func main() {
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

	// Setup Identity Module
	userRepo := identityInfra.NewPostgresUserRepository(dbPool)
	authSecret := cfg.App.Env // Using env as mock secret for now
	authService := identityApp.NewAuthService(userRepo, authSecret)
	identityHandler := identityHttp.NewHandler(authService)

	// Setup Learner Module
	learnerRepo := learnerInfra.NewPostgresLearnerProfileRepository(dbPool)
	updatePreferencesUseCase := learnerApp.NewUpdatePreferencesUseCase(learnerRepo)
	learnerHandler := learnerHttp.NewHandler(updatePreferencesUseCase)

	// Setup Knowledge Module
	knowledgeRepo := knowledgeInfra.NewPostgresKnowledgeRepository(dbPool)
	aiClientMock := &aiclient.MockClient{}
	knowledgeService := knowledgeApp.NewKnowledgeService(knowledgeRepo, aiClientMock)
	knowledgeHandler := knowledgeHttp.NewHandler(knowledgeService)

	// Setup Feedback Module
	feedbackRepo := feedbackInfra.NewPostgresFeedbackRepository(dbPool)
	recordFeedbackUseCase := feedbackApp.NewRecordResourceFeedbackUseCase(feedbackRepo)
	feedbackHandler := feedbackHttp.NewHandler(recordFeedbackUseCase)

	// Setup Goals Module
	goalRepo := infrastructure.NewPostgresGoalRepository(dbPool)
	createGoalUseCase := application.NewCreateGoalUseCase(goalRepo, aiClientMock, knowledgeService)
	getCurrentGoalUseCase := application.NewGetCurrentGoalUseCase(goalRepo)
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
	getSummaryUseCase := progressApp.NewGetProgressSummaryUseCase(progressRepo)
	mockGoalSvc := &mockGoalsService{}
	getGoalSummaryUseCase := progressApp.NewGetGoalCompletionSummaryUseCase(progressRepo, mockGoalSvc)
	progressHandler := progressHttp.NewHandler(getSummaryUseCase, getGoalSummaryUseCase, recordEngagementUseCase)

	// Setup Resources Module
	resourcesRepo := resourcesInfra.NewPostgresResourceRepository(dbPool)
	conceptValidator := &mockConceptValidationService{}
	createResourceUseCase := resourcesApp.NewCreateResourceUseCase(resourcesRepo, conceptValidator)
	updateResourceUseCase := resourcesApp.NewUpdateResourceUseCase(resourcesRepo)
	listResourcesUseCase := resourcesApp.NewListResourcesUseCase(resourcesRepo)
	getFeedbackSignalsUseCase := resourcesApp.NewGetFeedbackSignalsUseCase(resourcesRepo)
	getAlternateUseCase := resourcesApp.NewGetAlternateResourcesUseCase(resourcesRepo, aiClientMock)
	explainUseCase := resourcesApp.NewExplainResourceRelevanceUseCase(aiClientMock)
	resourcesHandler := resourcesHttp.NewHandler(createResourceUseCase, updateResourceUseCase, listResourcesUseCase, getFeedbackSignalsUseCase, getAlternateUseCase, explainUseCase)

	// Setup Assessment Module (Injecting real progress/evidence service instead of mock)
	assessmentRepo := assessmentInfra.NewPostgresAssessmentRepository(dbPool)
	getAssessmentUseCase := assessmentApp.NewGetAssessmentUseCase(assessmentRepo, knowledgeService)
	submitAssessmentUseCase := assessmentApp.NewSubmitAssessmentUseCase(assessmentRepo, knowledgeService, aiClientMock, recordEvidenceUseCase)
	assessmentHandler := assessmentHttp.NewHandler(getAssessmentUseCase, submitAssessmentUseCase)

	// Setup Roadmap Module
	roadmapRepo := roadmapInfra.NewPostgresRoadmapRepository(dbPool)
	roadmapGoalsMock := &mockRoadmapGoalsService{}
	roadmapKnowledgeMock := &mockRoadmapKnowledgeService{}
	roadmapCompetencyMock := &mockRoadmapCompetencyService{}
	roadmapResourcesMock := &mockRoadmapResourcesService{}
	roadmapAIMock := &mockRoadmapAIClientService{}
	regenerateRoadmapUseCase := roadmapApp.NewRegenerateRoadmapUseCase(
		txManager, roadmapRepo, roadmapGoalsMock, roadmapKnowledgeMock, roadmapCompetencyMock, roadmapResourcesMock, roadmapAIMock,
	)
	getActiveRoadmapUseCase := roadmapApp.NewGetActiveRoadmapUseCase(roadmapRepo)
	getConceptExplanationUseCase := roadmapApp.NewGetConceptExplanationUseCase(roadmapAIMock)
	roadmapHandler := roadmapHttp.NewHandler(getActiveRoadmapUseCase, regenerateRoadmapUseCase, getConceptExplanationUseCase)

	// Setup Projects Module
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
	// We instantiate the Redis bus here so other modules can publish to it.
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
	notificationsHandler.RegisterRoutes(mux)
	adminHandler.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

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

type mockGoalsService struct{}

func (m *mockGoalsService) GetCurrentGoalID(ctx context.Context, learnerID string) (string, error) {
	return "goal-mock-id", nil
}

func (m *mockGoalsService) GetGoalConceptCount(ctx context.Context, goalID string) (int, error) {
	return 5, nil
}

type mockConceptValidationService struct{}

func (m *mockConceptValidationService) ValidateConcepts(ctx context.Context, conceptIDs []string) error {
	return nil
}

type mockRoadmapGoalsService struct{}

func (m *mockRoadmapGoalsService) GetActiveGoal(ctx context.Context, learnerID string) (roadmapApp.Goal, error) {
	return roadmapApp.Goal{ID: "goal-id", KnowledgeStructureID: "ks-id", GoalText: "Learn stuff"}, nil
}

type mockRoadmapKnowledgeService struct{}

func (m *mockRoadmapKnowledgeService) ValidatePrerequisites(ctx context.Context, structureID string, orderedConceptIDs []string) error {
	return nil
}

type mockRoadmapCompetencyService struct{}

func (m *mockRoadmapCompetencyService) GetLearnerCompetencies(ctx context.Context, learnerID string) (map[string]string, error) {
	return map[string]string{}, nil
}

type mockRoadmapResourcesService struct{}

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

func (m *mockRoadmapResourcesService) ValidateResources(ctx context.Context, resourceIDs []string) error {
	return nil
}

type mockRoadmapAIClientService struct{}

func (m *mockRoadmapAIClientService) GenerateRoadmapProposal(ctx context.Context, req roadmapApp.AIProposalRequest) (roadmapApp.AIProposalResponse, error) {
	return roadmapApp.AIProposalResponse{
		Items: []roadmapApp.AIProposedItem{
			{ConceptID: "concept-1"},
			{ConceptID: "concept-2"},
		},
	}, nil
}
func (m *mockRoadmapAIClientService) GetConceptExplanation(ctx context.Context, conceptID string) (string, error) {
	return "because AI said so", nil
}
