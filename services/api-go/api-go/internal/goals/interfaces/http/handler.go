package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/goals/application"
)

type Handler struct {
	createGoalUseCase     *application.CreateGoalUseCase
	getCurrentGoalUseCase *application.GetCurrentGoalUseCase
}

func NewHandler(createGoal *application.CreateGoalUseCase, getCurrentGoal *application.GetCurrentGoalUseCase) *Handler {
	return &Handler{
		createGoalUseCase:     createGoal,
		getCurrentGoalUseCase: getCurrentGoal,
	}
}

type CreateGoalRequest struct {
	GoalText string `json:"goal_text"`
}

func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	// Simulated auth extraction (since identity middleware is not provided yet)
	// In reality, this would come from r.Context().Value("learner_id")
	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "bad request"}`, http.StatusBadRequest)
		return
	}

	goal, err := h.createGoalUseCase.Execute(r.Context(), learnerID, req.GoalText)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) GetCurrentGoal(w http.ResponseWriter, r *http.Request) {
	// Simulated auth extraction
	learnerID := r.Header.Get("X-Learner-ID")
	if learnerID == "" {
		learnerID = "default-learner-id"
	}

	goal, err := h.getCurrentGoalUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /goals", h.CreateGoal)
	mux.HandleFunc("GET /goals/current", h.GetCurrentGoal)
}
