package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/competency/application"
)

type Handler struct {
	getDetailUseCase  *application.GetCompetencyDetailUseCase
	getHistoryUseCase *application.GetCompetencyHistoryUseCase
}

func NewHandler(getDetail *application.GetCompetencyDetailUseCase, getHistory *application.GetCompetencyHistoryUseCase) *Handler {
	return &Handler{
		getDetailUseCase:  getDetail,
		getHistoryUseCase: getHistory,
	}
}

func (h *Handler) GetCompetencyDetail(w http.ResponseWriter, r *http.Request) {
	conceptID := r.URL.Query().Get("conceptId") // Temporary parse query, full spec may differ
	if conceptID == "" {
		http.Error(w, `{"error": "conceptId required"}`, http.StatusBadRequest)
		return
	}

	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	record, err := h.getDetailUseCase.Execute(r.Context(), learnerID, conceptID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func (h *Handler) GetCompetencyHistory(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("conceptId")
	if conceptID == "" {
		http.Error(w, `{"error": "conceptId required"}`, http.StatusBadRequest)
		return
	}

	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	history, err := h.getHistoryUseCase.Execute(r.Context(), learnerID, conceptID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"history": history})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// e.g. GET /competency/detail?conceptId=xxx
	mux.HandleFunc("GET /competency/detail", h.GetCompetencyDetail)
	mux.HandleFunc("GET /competency/{conceptId}/history", h.GetCompetencyHistory)
}
