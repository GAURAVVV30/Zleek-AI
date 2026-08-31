package http

import (
	"encoding/json"
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/notifications/application"
	"github.com/hcl-backend/services/api-go/internal/notifications/domain"
)

type Handler struct {
	getNotifications *application.GetNotificationsUseCase
	markRead         *application.MarkNotificationReadUseCase
}

func NewHandler(
	getNotifications *application.GetNotificationsUseCase,
	markRead *application.MarkNotificationReadUseCase,
) *Handler {
	return &Handler{
		getNotifications: getNotifications,
		markRead:         markRead,
	}
}

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	notifications, err := h.getNotifications.Execute(r.Context(), learnerID)
	if err != nil {
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if notifications == nil {
		notifications = []domain.Notification{} // Return empty array instead of null
	}
	json.NewEncoder(w).Encode(notifications)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	notificationID := r.PathValue("id")
	if notificationID == "" {
		http.Error(w, `{"error": "notification id required"}`, http.StatusBadRequest)
		return
	}

	if err := h.markRead.Execute(r.Context(), learnerID, notificationID); err != nil {
		if err == domain.ErrNotificationNotFound {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		if err == domain.ErrUnauthorized {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusForbidden)
			return
		}
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "read"})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /notifications", h.GetNotifications)
	mux.HandleFunc("PATCH /notifications/{id}/read", h.MarkRead)
}
