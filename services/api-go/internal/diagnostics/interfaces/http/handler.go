package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/diagnostics/application"
	"github.com/hcl-backend/services/api-go/internal/diagnostics/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
)

type Handler struct {
	startUseCase   *application.StartDiagnosticUseCase
	answerUseCase  *application.AnswerDiagnosticUseCase
	resultsUseCase *application.ResultsDiagnosticUseCase
}

func NewHandler(
	start *application.StartDiagnosticUseCase,
	answer *application.AnswerDiagnosticUseCase,
	results *application.ResultsDiagnosticUseCase,
) *Handler {
	return &Handler{startUseCase: start, answerUseCase: answer, resultsUseCase: results}
}

func (h *Handler) StartDiagnostic(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	res, err := h.startUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Start a goal before running the diagnostic")
		return
	}
	httpx.Envelope(w, http.StatusOK, res)
}

type answerRequest struct {
	QuestionID       string `json:"questionId"`
	SelectedOptionID string `json:"selectedOptionId"`
}

func (h *Handler) AnswerDiagnostic(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	sessionID := r.PathValue("sessionId")
	var body answerRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	res, err := h.answerUseCase.Execute(r.Context(), learnerID, sessionID, body.QuestionID, body.SelectedOptionID)
	if err != nil {
		switch err {
		case domain.ErrSessionNotFound:
			httpx.Error(w, http.StatusNotFound, "Diagnostic session not found")
		case domain.ErrSessionComplete:
			httpx.Error(w, http.StatusGone, "Diagnostic already completed")
		default:
			httpx.Error(w, http.StatusBadRequest, "Invalid answer")
		}
		return
	}
	httpx.Envelope(w, http.StatusOK, res)
}

func (h *Handler) GetResults(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	sessionID := r.PathValue("sessionId")
	results, err := h.resultsUseCase.Execute(r.Context(), learnerID, sessionID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "Diagnostic session not found")
		return
	}
	httpx.Envelope(w, http.StatusOK, results)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /diagnostics/start", h.StartDiagnostic)
	mux.HandleFunc("POST /diagnostics/{sessionId}/answer", h.AnswerDiagnostic)
	mux.HandleFunc("GET /diagnostics/{sessionId}/results", h.GetResults)
}
