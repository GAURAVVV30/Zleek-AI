package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
	"github.com/hcl-backend/services/api-go/internal/progress/application"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type Handler struct {
	recordEvidence     UseCase
	recordEngagement   *application.RecordEngagementUseCase
	getSummary         *application.GetProgressSummaryUseCase
	getGoalSummary     *application.GetGoalCompletionSummaryUseCase
	getCompletionBadge *application.GetCompletionBadgeUseCase
}

type UseCase interface {
	RecordEvidence(ctx context.Context, evidence *domain.Evidence) (string, error)
}

func NewHandler(
	recordEvidence UseCase,
	recordEngagement *application.RecordEngagementUseCase,
	getSummary *application.GetProgressSummaryUseCase,
	getGoalSummary *application.GetGoalCompletionSummaryUseCase,
	getCompletionBadge *application.GetCompletionBadgeUseCase,
) *Handler {
	return &Handler{
		recordEvidence:     recordEvidence,
		recordEngagement:   recordEngagement,
		getSummary:         getSummary,
		getGoalSummary:     getGoalSummary,
		getCompletionBadge: getCompletionBadge,
	}
}

type engagementRequest struct {
	ConceptID string `json:"conceptId"`
	Action    string `json:"action"`
}

func (h *Handler) RecordEngagement(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	conceptID := r.PathValue("conceptId")
	var body engagementRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body.Action = r.URL.Query().Get("action")
	}
	if conceptID == "" {
		conceptID = body.ConceptID
	}
	if err := h.recordEngagement.RecordEngagement(r.Context(), learnerID, conceptID, body.Action); err != nil {
		if errors.Is(err, domain.ErrPrerequisiteNotMet) {
			httpx.Error(w, http.StatusConflict, "Previous module must be completed first")
			return
		}
		httpx.Error(w, http.StatusUnprocessableEntity, "Invalid engagement event")
		return
	}
	httpx.Envelope(w, http.StatusCreated, map[string]any{"status": "recorded"})
}

func (h *Handler) GetProgressSummary(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	summary, err := h.getSummary.Execute(r.Context(), learnerID)
	if err != nil {
		if err == domain.ErrNoActivePath {
			httpx.Envelope(w, http.StatusOK, &domain.Summary{
				OverallCompletionPercentage: 0,
				TotalConcepts:               0,
				CompletedConcepts:           0,
				ActiveRemediations:          0,
				CompetencyBreakdown:         []domain.SummaryRow{},
			})
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "Failed to load progress summary")
		return
	}
	httpx.Envelope(w, http.StatusOK, summary)
}

func (h *Handler) GetGoalSummary(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	summary, err := h.getGoalSummary.Execute(r.Context(), learnerID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to load goal summary")
		return
	}
	httpx.Envelope(w, http.StatusOK, summary)
}

func (h *Handler) GetCompletionBadge(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	badgeRes, err := h.getCompletionBadge.Execute(r.Context(), learnerID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to evaluate completion badge eligibility")
		return
	}

	if !badgeRes.Eligible {
		httpx.Envelope(w, http.StatusForbidden, badgeRes)
		return
	}

	httpx.Envelope(w, http.StatusOK, badgeRes)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /concepts/{conceptId}/engagement", h.RecordEngagement)
	mux.HandleFunc("GET /progress/summary", h.GetProgressSummary)
	mux.HandleFunc("GET /progress/goal-summary", h.GetGoalSummary)
	mux.HandleFunc("GET /goals/current/completion-summary", h.GetGoalSummary)
	mux.HandleFunc("GET /goals/completion-summary", h.GetGoalSummary)
	mux.HandleFunc("GET /api/v1/goals/current/completion-summary", h.GetGoalSummary)
	mux.HandleFunc("GET /api/v1/progress/goal-summary", h.GetGoalSummary)
	mux.HandleFunc("GET /progress/completion-badge", h.GetCompletionBadge)
	mux.HandleFunc("GET /api/v1/progress/completion-badge", h.GetCompletionBadge)
	mux.HandleFunc("GET /completion-badge", h.GetCompletionBadge)
}
