package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/progress/application"
)

type Handler struct {
	getSummaryUseCase               *application.GetProgressSummaryUseCase
	getGoalCompletionSummaryUseCase *application.GetGoalCompletionSummaryUseCase
	recordEngagementUseCase         *application.RecordEngagementUseCase
}

func NewHandler(
	getSummary *application.GetProgressSummaryUseCase,
	getGoalSummary *application.GetGoalCompletionSummaryUseCase,
	recordEngagement *application.RecordEngagementUseCase,
) *Handler {
	return &Handler{
		getSummaryUseCase:               getSummary,
		getGoalCompletionSummaryUseCase: getGoalSummary,
		recordEngagementUseCase:         recordEngagement,
	}
}

func (h *Handler) GetProgressSummary(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	summary, err := h.getSummaryUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *Handler) GetGoalCompletionSummary(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	summary, err := h.getGoalCompletionSummaryUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *Handler) RecordEngagement(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("id")
	if conceptID == "" {
		http.Error(w, `{"error": "concept id required"}`, http.StatusBadRequest)
		return
	}

	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	var payload struct {
		EventType string `json:"event_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error": "bad request: invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	// The frontend uses /concepts/{id}/engagement but we need pathItemID for the event.
	// For now we map conceptID to pathItemID to satisfy the contract.
	if err := h.recordEngagementUseCase.Execute(r.Context(), learnerID, conceptID, payload.EventType); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /progress/summary", h.GetProgressSummary)
	mux.HandleFunc("GET /goals/current/completion-summary", h.GetGoalCompletionSummary)
	mux.HandleFunc("POST /concepts/{id}/engagement", h.RecordEngagement)
}
