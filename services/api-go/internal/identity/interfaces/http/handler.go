package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/identity/application"
	"github.com/hcl-backend/services/api-go/internal/identity/domain"
)

type Handler struct {
	authService *application.AuthService
}

func NewHandler(authService *application.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", h.handleSignup)
	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
	mux.HandleFunc("GET /auth/me", h.handleGetMe)
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

func (h *Handler) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		h.errorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	resp, err := h.authService.Signup(r.Context(), req.Email, req.Password, domain.RoleLearner)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			h.errorResponse(w, http.StatusConflict, err.Error())
			return
		}
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.successResponse(w, http.StatusCreated, resp)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			h.errorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.successResponse(w, http.StatusOK, resp)
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			h.errorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.successResponse(w, http.StatusOK, resp)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	// For stateless JWT, logout is often handled client side or with a token blacklist.
	// Contract says return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			h.errorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"role":  string(user.Role),
	})
}

func (h *Handler) successResponse(w http.ResponseWriter, status int, data *application.AuthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(authResponse{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}{
			ID:    data.User.ID,
			Email: data.User.Email,
			Role:  string(data.User.Role),
		},
	})
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    status,
			"message": message,
		},
	})
}
