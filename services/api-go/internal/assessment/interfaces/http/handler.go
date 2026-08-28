package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/assessment/application"
)

type Handler struct {
	getAssessmentUseCase    *application.GetAssessmentUseCase
	submitAssessmentUseCase *application.SubmitAssessmentUseCase
}

func NewHandler(getAssessment *application.GetAssessmentUseCase, submitAssessment *application.SubmitAssessmentUseCase) *Handler {
	return &Handler{
		getAssessmentUseCase:    getAssessment,
		submitAssessmentUseCase: submitAssessment,
	}
}

func (h *Handler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	// Simple path param extraction since we don't have a robust router in this scaffold yet.
	// We simulate the conceptID extraction.
	conceptID := r.URL.Query().Get("id") // Assume /concepts/assessment?id=... for simple mounting, or custom parsing
	if conceptID == "" {
		// Try to extract from path like /concepts/{id}/assessment
		// Very naive extraction:
		// Example: /concepts/123/assessment
		// Parts: "", "concepts", "123", "assessment"
		// For the sake of this implementation we assume standard mounting or path value if Go 1.22+ is used.
		conceptID = r.PathValue("id")
	}

	if conceptID == "" {
		http.Error(w, `{"error": "concept id required"}`, http.StatusBadRequest)
		return
	}

	assessment, err := h.getAssessmentUseCase.Execute(r.Context(), conceptID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(assessment)
}

func (h *Handler) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("id")
	if conceptID == "" {
		http.Error(w, `{"error": "concept id required"}`, http.StatusBadRequest)
		return
	}

	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	var submission json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, `{"error": "bad request: invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	result, err := h.submitAssessmentUseCase.Execute(r.Context(), learnerID, conceptID, submission)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Uses Go 1.22+ path values
	mux.HandleFunc("GET /concepts/{id}/assessment", h.GetAssessment)
	mux.HandleFunc("POST /concepts/{id}/assessment/submit", h.SubmitAssessment)
}
