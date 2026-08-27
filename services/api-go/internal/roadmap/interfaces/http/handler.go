package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/roadmap/application"
)

type Handler struct {
	getUseCase         *application.GetActiveRoadmapUseCase
	regenerateUseCase  *application.RegenerateRoadmapUseCase
	explanationUseCase *application.GetConceptExplanationUseCase
}

func NewHandler(
	get *application.GetActiveRoadmapUseCase,
	regenerate *application.RegenerateRoadmapUseCase,
	explanation *application.GetConceptExplanationUseCase,
) *Handler {
	return &Handler{
		getUseCase:         get,
		regenerateUseCase:  regenerate,
		explanationUseCase: explanation,
	}
}

func (h *Handler) GetActiveRoadmap(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = "default-learner-id" // For testing without middleware
	}

	path, items, err := h.getUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"path":       path,
		"path_items": items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) RegenerateRoadmap(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	path, items, err := h.regenerateUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"path":       path,
		"path_items": items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetConceptExplanation(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("conceptId")
	if conceptID == "" {
		http.Error(w, `{"error": "conceptId required"}`, http.StatusBadRequest)
		return
	}

	explanation, err := h.explanationUseCase.Execute(r.Context(), conceptID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"explanation": explanation})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /roadmap", h.GetActiveRoadmap)
	mux.HandleFunc("POST /roadmap/regenerate", h.RegenerateRoadmap)
	mux.HandleFunc("GET /roadmap/concepts/{conceptId}/why", h.GetConceptExplanation)
}
