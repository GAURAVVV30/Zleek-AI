package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/admin/application"
	"github.com/hcl-backend/services/api-go/internal/admin/domain"
)

type Handler struct {
	listUsers    *application.ListUsersUseCase
	updateUser   *application.UpdateUserUseCase
	getAuditLog  *application.GetAuditLogUseCase
}

func NewHandler(
	listUsers *application.ListUsersUseCase,
	updateUser *application.UpdateUserUseCase,
	getAuditLog *application.GetAuditLogUseCase,
) *Handler {
	return &Handler{
		listUsers:   listUsers,
		updateUser:  updateUser,
		getAuditLog: getAuditLog,
	}
}

func (h *Handler) isAdmin(r *http.Request) bool {
	// Simple mock authorization logic for this module.
	// In production, this would rely on the JWT context injected by Identity middleware.
	return r.Header.Get("X-User-Role") == "admin"
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		return
	}

	users, err := h.listUsers.Execute(r.Context())
	if err != nil {
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if users == nil {
		users = []domain.User{} // return [] instead of null
	}
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		return
	}

	adminID := r.Header.Get("X-User-ID")
	if adminID == "" {
		adminID = "system-admin-id" // mock fallback for testing
	}

	// According to yaml, it's /admin/users without an ID in the path. Let's decode from body.
	
	type requestBody struct {
		ID     string  `json:"id"`
		Role   *string `json:"role"`
		Status *string `json:"status"`
	}

	var req requestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid payload"}`, http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.updateUser.Execute(r.Context(), adminID, req.ID, req.Role, req.Status)
	if err != nil {
		if err == domain.ErrUserNotFound {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		if err == domain.ErrInvalidRole || err == domain.ErrInvalidStatus || err == domain.ErrSelfDemotion {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		return
	}

	records, err := h.getAuditLog.Execute(r.Context())
	if err != nil {
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if records == nil {
		records = []domain.AuditRecord{} // return [] instead of null
	}
	json.NewEncoder(w).Encode(records)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/users", h.ListUsers)
	mux.HandleFunc("PATCH /admin/users", h.UpdateUser)
	mux.HandleFunc("GET /admin/audit-log", h.GetAuditLog)
}
