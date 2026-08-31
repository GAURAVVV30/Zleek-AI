package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/learner/application"
)

type Handler struct {
	updatePreferencesUseCase *application.UpdatePreferencesUseCase
}

func NewHandler(updatePreferencesUseCase *application.UpdatePreferencesUseCase) *Handler {
	return &Handler{
		updatePreferencesUseCase: updatePreferencesUseCase,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("PATCH /profile/preferences", h.handlePatchPreferences)
}

func (h *Handler) handlePatchPreferences(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	profile, err := h.updatePreferencesUseCase.Execute(r.Context(), learnerID, req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile.Preferences)
}
