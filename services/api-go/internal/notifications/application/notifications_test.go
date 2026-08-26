package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/hcl-backend/services/api-go/internal/notifications/application"
	"github.com/hcl-backend/services/api-go/internal/notifications/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/events"
)

type mockRepo struct {
	notifications map[string]*domain.Notification
}

func (m *mockRepo) Create(ctx context.Context, n *domain.Notification) error {
	n.ID = "n-1"
	m.notifications[n.ID] = n
	return nil
}

func (m *mockRepo) ListForUser(ctx context.Context, learnerID string) ([]domain.Notification, error) {
	var res []domain.Notification
	for _, n := range m.notifications {
		if n.LearnerID == learnerID {
			res = append(res, *n)
		}
	}
	return res, nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	if n, ok := m.notifications[id]; ok {
		return n, nil
	}
	return nil, domain.ErrNotificationNotFound
}

func (m *mockRepo) MarkRead(ctx context.Context, id string) error {
	if n, ok := m.notifications[id]; ok {
		now := time.Now()
		n.ReadAt = &now
		return nil
	}
	return domain.ErrNotificationNotFound
}

func TestGetNotifications(t *testing.T) {
	repo := &mockRepo{notifications: make(map[string]*domain.Notification)}
	repo.notifications["1"] = &domain.Notification{ID: "1", LearnerID: "user-1"}
	repo.notifications["2"] = &domain.Notification{ID: "2", LearnerID: "user-2"}

	uc := application.NewGetNotificationsUseCase(repo)
	ns, _ := uc.Execute(context.Background(), "user-1")
	if len(ns) != 1 || ns[0].ID != "1" {
		t.Fatal("wrong notifications returned")
	}
}

func TestMarkRead(t *testing.T) {
	repo := &mockRepo{notifications: make(map[string]*domain.Notification)}
	repo.notifications["1"] = &domain.Notification{ID: "1", LearnerID: "user-1"}

	uc := application.NewMarkNotificationReadUseCase(repo)

	// wrong user
	err := uc.Execute(context.Background(), "user-2", "1")
	if err != domain.ErrUnauthorized {
		t.Fatal("expected unauthorized")
	}

	// right user
	err = uc.Execute(context.Background(), "user-1", "1")
	if err != nil {
		t.Fatal(err)
	}

	if repo.notifications["1"].ReadAt == nil {
		t.Fatal("expected read_at to be set")
	}
}

func TestEventHandler(t *testing.T) {
	repo := &mockRepo{notifications: make(map[string]*domain.Notification)}
	handler := application.NewEventHandler(repo)

	e := events.Event{
		Type: events.EventTypeCompetencyUpdated,
		Payload: events.CompetencyUpdatedPayload{
			LearnerID: "user-1",
			ConceptID: "c-1",
			NewState:  "competent",
		},
	}

	err := handler.HandleCompetencyUpdated(context.Background(), e)
	if err != nil {
		t.Fatal(err)
	}

	if len(repo.notifications) != 1 {
		t.Fatal("expected 1 notification created")
	}
}
