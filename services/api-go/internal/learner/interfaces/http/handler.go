package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/learner/application"
	"github.com/hcl-backend/services/api-go/internal/learner/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
)

type Handler struct {
	updatePreferencesUseCase *application.UpdatePreferencesUseCase
	updateSettingsUseCase    *application.UpdateSettingsUseCase
}

func NewHandler(
	updatePreferences *application.UpdatePreferencesUseCase,
	updateSettings *application.UpdateSettingsUseCase,
) *Handler {
	return &Handler{
		updatePreferencesUseCase: updatePreferences,
		updateSettingsUseCase:    updateSettings,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("PATCH /profile/preferences", h.handlePatchPreferences)
	mux.HandleFunc("PATCH /profile/settings", h.handlePatchSettings)
}

type preferencesRequest struct {
	WeeklyHours     string   `json:"weeklyHours"`
	PreferredFormat []string `json:"preferredFormat"`
	ExperienceLevel string   `json:"experienceLevel"`
	Gender          string   `json:"gender,omitempty"`
	AvatarURL       string   `json:"avatarUrl,omitempty"`
}

func (h *Handler) handlePatchPreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req preferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	for _, f := range req.PreferredFormat {
		if !domain.ValidFormat(f) {
			httpx.Error(w, http.StatusBadRequest, "preferredFormat only accepts video, article, interactive")
			return
		}
	}
	if req.WeeklyHours != "" && !domain.ValidAvailability(req.WeeklyHours) {
		httpx.Error(w, http.StatusBadRequest, "weeklyHours only accepts lt_5, 5_10, 10_20, gt_20")
		return
	}
	if req.ExperienceLevel != "" && !domain.ValidExperience(req.ExperienceLevel) {
		httpx.Error(w, http.StatusBadRequest, "experienceLevel only accepts beginner, intermediate, advanced")
		return
	}

	prefs, err := h.updatePreferencesUseCase.Execute(r.Context(), userID, application.PreferencesInput{
		TimeAvailability: req.WeeklyHours,
		FormatPreference: req.PreferredFormat,
		ExperienceLevel:  req.ExperienceLevel,
		Gender:           req.Gender,
		AvatarURL:        req.AvatarURL,
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to save preferences")
		return
	}

	httpx.Envelope(w, http.StatusOK, map[string]any{
		"message":         "Preferences saved.",
		"weeklyHours":     prefs.TimeAvailability,
		"preferredFormat": splitFormats(prefs.FormatPreference),
		"experienceLevel": prefs.PriorExperience,
	})
}

type settingsRequest struct {
	FullName string `json:"fullName"`
	Timezone string `json:"timezone"`
	Theme    string `json:"theme"`
}

func (h *Handler) handlePatchSettings(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req settingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.updateSettingsUseCase.Execute(r.Context(), userID, req.FullName, req.Timezone, req.Theme); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Failed to save settings")
		return
	}
	httpx.Envelope(w, http.StatusOK, settingsRequest{
		FullName: req.FullName,
		Timezone: req.Timezone,
		Theme:    req.Theme,
	})
}

func splitFormats(v string) []string {
	if v == "" {
		return []string{}
	}
	out := []string{}
	cur := ""
	for i := 0; i < len(v); i++ {
		if v[i] == ',' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		cur += string(v[i])
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
