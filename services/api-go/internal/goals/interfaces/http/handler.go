package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/goals/application"
	"github.com/hcl-backend/services/api-go/internal/goals/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
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

type createGoalRequest struct {
	GoalText string `json:"goalText"`
}

// GoalView is the goals/current shape the frontend consumes.
type GoalView struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	DomainID   string `json:"domainId"`
	DomainName string `json:"domainName"`
	Status     string `json:"status"`
}

func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GoalText == "" {
		httpx.Error(w, http.StatusBadRequest, "goalText is required")
		return
	}

	mapping, err := h.createGoalUseCase.Execute(r.Context(), learnerID, req.GoalText)
	if err != nil {
		if errors.Is(err, domain.ErrAIProposalInvalid) {
			httpx.Error(w, http.StatusUnprocessableEntity, "Goal could not be mapped to a supported role")
			return
		}
		if errors.Is(err, domain.ErrGoalNotFound) || errors.Is(err, domain.ErrKnowledgeUnpublished) {
			httpx.Error(w, http.StatusUnprocessableEntity, "Goal could not be mapped to a published knowledge structure")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "Failed to create goal")
		return
	}
	httpx.Envelope(w, http.StatusCreated, mapping)
}

func (h *Handler) GetCurrentGoal(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	goal, err := h.getCurrentGoalUseCase.Execute(r.Context(), learnerID)
	if err != nil {
		if errors.Is(err, domain.ErrGoalNotFound) {
			httpx.Envelope(w, http.StatusOK, GoalView{})
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "Failed to load goal")
		return
	}
	httpx.Envelope(w, http.StatusOK, GoalView{
		ID:         goal.ID,
		Title:      goal.GoalText,
		DomainID:   goal.DomainID,
		DomainName: goal.DomainName,
		Status:     goal.Status,
	})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /goals", h.CreateGoal)
	mux.HandleFunc("GET /goals/current", h.GetCurrentGoal)
}
