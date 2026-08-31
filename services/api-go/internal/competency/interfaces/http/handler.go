package http

import (
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/competency/application"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
)

type Handler struct {
	getDetailUseCase  *application.GetCompetencyDetailUseCase
	getHistoryUseCase *application.GetCompetencyHistoryUseCase
}

func NewHandler(getDetail *application.GetCompetencyDetailUseCase, getHistory *application.GetCompetencyHistoryUseCase) *Handler {
	return &Handler{getDetailUseCase: getDetail, getHistoryUseCase: getHistory}
}

func (h *Handler) GetCompetencyDetail(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	records, err := h.getDetailUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to load competency records")
		return
	}
	httpx.Envelope(w, http.StatusOK, records)
}

func (h *Handler) GetCompetencyHistory(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	conceptID := r.PathValue("conceptId")
	if conceptID == "" {
		httpx.Error(w, http.StatusBadRequest, "conceptId required")
		return
	}
	history, err := h.getHistoryUseCase.Execute(r.Context(), learnerID, conceptID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to load competency history")
		return
	}
	httpx.Envelope(w, http.StatusOK, history)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /competency/detail", h.GetCompetencyDetail)
	mux.HandleFunc("GET /competency/{conceptId}/history", h.GetCompetencyHistory)
}
