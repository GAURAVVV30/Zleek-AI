package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/identity/application"
	"github.com/hcl-backend/services/api-go/internal/identity/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
)

type Handler struct {
	authService *application.AuthService
}

func NewHandler(authService *application.AuthService) *Handler {
	return &Handler{authService: authService}
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", h.handleSignup)
	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
	mux.HandleFunc("GET /auth/me", h.handleGetMe)
	mux.HandleFunc("POST /auth/change-password", h.handleChangePassword)
}

func (h *Handler) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		httpx.Error(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	resp, err := h.authService.Signup(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrWeakPassword):
			httpx.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrUserAlreadyExists):
			httpx.Error(w, http.StatusConflict, err.Error())
		default:
			httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	httpx.Envelope(w, http.StatusCreated, authResponsePayload(resp))
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httpx.Error(w, http.StatusUnauthorized, err.Error())
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	httpx.Envelope(w, http.StatusOK, authResponsePayload(resp))
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}
	httpx.Envelope(w, http.StatusOK, authResponsePayload(resp))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if actorID := r.Header.Get("X-User-ID"); actorID != "" {
		_ = h.authService.Logout(r.Context(), actorID)
	}
	httpx.NoContent(w)
}

func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			httpx.Error(w, http.StatusNotFound, "User not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	httpx.Envelope(w, http.StatusOK, userPayload(user))
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		httpx.Error(w, http.StatusBadRequest, "currentPassword and newPassword are required")
		return
	}
	if err := h.authService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			httpx.Error(w, http.StatusBadRequest, "Current password is incorrect")
		case errors.Is(err, domain.ErrWeakPassword):
			httpx.Error(w, http.StatusBadRequest, err.Error())
		default:
			httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	httpx.NoContent(w)
}

// mapUser converts a domain user into the JSON shape the React client reads.
func mapUser(u *domain.User) map[string]any {
	return map[string]any{
		"id":        u.ID,
		"email":     u.Email,
		"fullName":  u.FullName,
		"role":      string(u.Role),
		"status":    string(u.Status),
		"timezone":  u.Timezone,
		"theme":     u.Theme,
		"createdAt": u.CreatedAt,
	}
}

func userPayload(u *domain.User) map[string]any {
	return mapUser(u)
}

func authResponsePayload(resp *application.AuthResponse) map[string]any {
	return map[string]any{
		"accessToken":  resp.AccessToken,
		"refreshToken": resp.RefreshToken,
		"user":         mapUser(&resp.User),
	}
}
