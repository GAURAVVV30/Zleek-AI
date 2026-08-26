package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/feedback/application"
	"github.com/hcl-backend/services/api-go/internal/feedback/domain"
)

type Handler struct {
	recordResourceFeedbackUseCase *application.RecordResourceFeedbackUseCase
}

func NewHandler(recordResourceFeedbackUseCase *application.RecordResourceFeedbackUseCase) *Handler {
	return &Handler{
		recordResourceFeedbackUseCase: recordResourceFeedbackUseCase,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /resources/", h.handleResourceFeedback) // POST /resources/{resourceId}/feedback
}

type feedbackRequest struct {
	Rating  float64 `json:"rating"`
	Comment string  `json:"comment"`
}

func (h *Handler) handleResourceFeedback(w http.ResponseWriter, r *http.Request) {
	// extract /resources/{resourceId}/feedback
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 || parts[3] != "feedback" {
		http.NotFound(w, r)
		return
	}
	resourceID := parts[2]

	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	record, err := h.recordResourceFeedbackUseCase.Execute(r.Context(), learnerID, resourceID, req.Rating, req.Comment)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFeedback) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}
