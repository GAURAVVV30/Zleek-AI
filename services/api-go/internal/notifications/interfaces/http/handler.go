package http

import (
	"net/http"

	"github.com/hcl-backend/services/api-go/internal/notifications/application"
	"github.com/hcl-backend/services/api-go/internal/notifications/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
)

type Handler struct {
	getNotifications *application.GetNotificationsUseCase
	markRead         *application.MarkNotificationReadUseCase
}

func NewHandler(
	getNotifications *application.GetNotificationsUseCase,
	markRead *application.MarkNotificationReadUseCase,
) *Handler {
	return &Handler{getNotifications: getNotifications, markRead: markRead}
}

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	notifications, err := h.getNotifications.Execute(r.Context(), learnerID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if notifications == nil {
		notifications = []domain.Notification{}
	}
	httpx.Envelope(w, http.StatusOK, notifications)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	learnerID := r.Header.Get("X-User-ID")
	if learnerID == "" {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	notificationID := r.PathValue("id")
	if notificationID == "" {
		httpx.Error(w, http.StatusBadRequest, "Notification id required")
		return
	}
	if err := h.markRead.Execute(r.Context(), learnerID, notificationID); err != nil {
		switch err {
		case domain.ErrNotificationNotFound:
			httpx.Error(w, http.StatusNotFound, "Notification not found")
		case domain.ErrUnauthorized:
			httpx.Error(w, http.StatusForbidden, "Forbidden")
		default:
			httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	httpx.Envelope(w, http.StatusOK, map[string]string{"status": "read"})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /notifications", h.GetNotifications)
	mux.HandleFunc("PATCH /notifications/{id}/read", h.MarkRead)
}
