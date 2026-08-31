package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/assessment/application"
	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
)

type Handler struct {
	getUseCase    *application.GetAssessmentUseCase
	submitUseCase *application.SubmitAssessmentUseCase
}

func NewHandler(getUseCase *application.GetAssessmentUseCase, submitUseCase *application.SubmitAssessmentUseCase) *Handler {
	return &Handler{getUseCase: getUseCase, submitUseCase: submitUseCase}
}

func (h *Handler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("conceptId")
	if conceptID == "" {
		httpx.Error(w, http.StatusBadRequest, "conceptId required")
		return
	}
	quiz, err := h.getUseCase.Execute(r.Context(), conceptID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "Assessment not found for concept")
		return
	}
	httpx.Envelope(w, http.StatusOK, quiz)
}

func (h *Handler) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("conceptId")
	if conceptID == "" {
		httpx.Error(w, http.StatusBadRequest, "conceptId required")
		return
	}
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var submission json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	result, err := h.submitUseCase.Execute(r.Context(), learnerID, conceptID, submission)
	if err != nil {
		switch err {
		case domain.ErrInvalidSubmission:
			httpx.Error(w, http.StatusBadRequest, "Invalid submission payload")
		case domain.ErrConceptNotFound, domain.ErrAssessmentNotFound:
			httpx.Error(w, http.StatusNotFound, "Concept or assessment not found")
		case domain.ErrAIUnavailable:
			httpx.Error(w, http.StatusServiceUnavailable, "AI evaluation unavailable")
		default:
			httpx.Error(w, http.StatusBadRequest, "Evaluation failed")
		}
		return
	}
	httpx.Envelope(w, http.StatusOK, result)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /concepts/{conceptId}/assessment", h.GetAssessment)
	mux.HandleFunc("POST /concepts/{conceptId}/assessment/submit", h.SubmitAssessment)
}
