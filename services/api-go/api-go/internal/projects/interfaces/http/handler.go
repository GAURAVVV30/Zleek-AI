package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/projects/application"
	"github.com/hcl-backend/services/api-go/internal/projects/domain"
)

type Handler struct {
	getProject       *application.GetProjectUseCase
	submitProject    *application.SubmitProjectUseCase
	getProjectStatus *application.GetProjectStatusUseCase
}

func NewHandler(
	getProject *application.GetProjectUseCase,
	submitProject *application.SubmitProjectUseCase,
	getProjectStatus *application.GetProjectStatusUseCase,
) *Handler {
	return &Handler{
		getProject:       getProject,
		submitProject:    submitProject,
		getProjectStatus: getProjectStatus,
	}
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("id")
	if conceptID == "" {
		http.Error(w, `{"error": "concept id required"}`, http.StatusBadRequest)
		return
	}

	project, err := h.getProject.Execute(r.Context(), conceptID)
	if err != nil {
		if err == domain.ErrNoProjectForConcept {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) SubmitProject(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("id")
	if conceptID == "" {
		http.Error(w, `{"error": "concept id required"}`, http.StatusBadRequest)
		return
	}

	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	var meta domain.SubmissionMetadata
	if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
		http.Error(w, `{"error": "invalid payload"}`, http.StatusBadRequest)
		return
	}

	if meta.ArtifactReference == "" {
		http.Error(w, `{"error": "artifact reference required"}`, http.StatusBadRequest)
		return
	}

	if err := h.submitProject.Execute(r.Context(), learnerID, conceptID, meta); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "submitted"})
}

func (h *Handler) GetProjectStatus(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("id")
	if conceptID == "" {
		http.Error(w, `{"error": "concept id required"}`, http.StatusBadRequest)
		return
	}

	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	state, err := h.getProjectStatus.Execute(r.Context(), learnerID, conceptID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if state == nil {
		http.Error(w, `{"error": "no submission found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /concepts/{id}/project", h.GetProject)
	mux.HandleFunc("POST /concepts/{id}/project/submit", h.SubmitProject)
	mux.HandleFunc("GET /concepts/{id}/project/status", h.GetProjectStatus)
}
