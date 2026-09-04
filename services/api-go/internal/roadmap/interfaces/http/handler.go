package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
	"github.com/hcl-backend/services/api-go/internal/roadmap/application"
	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

type Handler struct {
	getUseCase         *application.GetActiveRoadmapUseCase
	regenerateUseCase  *application.RegenerateRoadmapUseCase
	explanationUseCase *application.GetConceptExplanationUseCase
	getTasksUseCase    *application.GetDailyTasksUseCase
	toggleTaskUseCase  *application.ToggleDailyTaskUseCase
}

func NewHandler(
	get *application.GetActiveRoadmapUseCase,
	regenerate *application.RegenerateRoadmapUseCase,
	explanation *application.GetConceptExplanationUseCase,
	getTasks *application.GetDailyTasksUseCase,
	toggleTask *application.ToggleDailyTaskUseCase,
) *Handler {
	return &Handler{
		getUseCase:         get,
		regenerateUseCase:  regenerate,
		explanationUseCase: explanation,
		getTasksUseCase:    getTasks,
		toggleTaskUseCase:  toggleTask,
	}
}

func (h *Handler) GetActiveRoadmap(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = r.Header.Get("X-Learner-ID")
	}
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	roadmap, err := h.getUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		// Auto-generate active roadmap from active goal if missing
		roadmap, err = h.regenerateUseCase.Execute(r.Context(), learnerID)
	}
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "No active roadmap")
		return
	}
	httpx.Envelope(w, http.StatusOK, roadmap)
}

func (h *Handler) RegenerateRoadmap(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = r.Header.Get("X-Learner-ID")
	}
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	roadmap, err := h.regenerateUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		if err == domain.ErrNoActiveGoal {
			httpx.Error(w, http.StatusBadRequest, "No active goal")
			return
		}
		httpx.Error(w, http.StatusBadRequest, "Roadmap regeneration failed. Check that your goal is active.")
		return
	}
	httpx.Envelope(w, http.StatusOK, roadmap)
}

func (h *Handler) GetConceptExplanation(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("conceptId")
	if conceptID == "" {
		httpx.Error(w, http.StatusBadRequest, "conceptId required")
		return
	}
	reason, err := h.explanationUseCase.Execute(r.Context(), conceptID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Explanation unavailable")
		return
	}
	httpx.Envelope(w, http.StatusOK, reason)
}

func (h *Handler) GetDailyTasks(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = r.Header.Get("X-Learner-ID")
	}
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	tasks, err := h.getTasksUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.Envelope(w, http.StatusOK, tasks)
}

func (h *Handler) ToggleDailyTask(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		learnerID = r.Header.Get("X-Learner-ID")
	}
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	taskID := r.PathValue("taskId")
	if taskID == "" {
		httpx.Error(w, http.StatusBadRequest, "taskId required")
		return
	}
	var payload struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	err := h.toggleTaskUseCase.Execute(r.Context(), learnerID, taskID, payload.Completed)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.Envelope(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /roadmap", h.GetActiveRoadmap)
	mux.HandleFunc("GET /api/v1/roadmap", h.GetActiveRoadmap)
	mux.HandleFunc("POST /roadmap/regenerate", h.RegenerateRoadmap)
	mux.HandleFunc("POST /api/v1/roadmap/regenerate", h.RegenerateRoadmap)
	mux.HandleFunc("GET /roadmap/concepts/{conceptId}/why", h.GetConceptExplanation)
	mux.HandleFunc("GET /api/v1/roadmap/concepts/{conceptId}/why", h.GetConceptExplanation)
	mux.HandleFunc("GET /roadmap/tasks", h.GetDailyTasks)
	mux.HandleFunc("GET /api/v1/roadmap/tasks", h.GetDailyTasks)
	mux.HandleFunc("POST /roadmap/tasks/{taskId}/toggle", h.ToggleDailyTask)
	mux.HandleFunc("POST /api/v1/roadmap/tasks/{taskId}/toggle", h.ToggleDailyTask)
}
