package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hcl-backend/services/api-go/internal/notifications/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/events"
)

type EventHandler struct {
	repo NotificationRepository
}

func NewEventHandler(repo NotificationRepository) *EventHandler {
	return &EventHandler{repo: repo}
}

func (h *EventHandler) HandleCompetencyUpdated(ctx context.Context, e events.Event) error {
	var payload events.CompetencyUpdatedPayload
	if err := extractPayload(e.Payload, &payload); err != nil {
		return err
	}

	notif := &domain.Notification{
		LearnerID: payload.LearnerID,
		EventType: events.EventTypeCompetencyUpdated,
		Payload:   marshalPayload(map[string]string{"message": "Your competency for concept " + payload.ConceptID + " is now " + payload.NewState}),
		CreatedAt: time.Now(),
	}

	return h.repo.Create(ctx, notif)
}

func (h *EventHandler) HandleConceptWeak(ctx context.Context, e events.Event) error {
	var payload events.ConceptWeakPayload
	if err := extractPayload(e.Payload, &payload); err != nil {
		return err
	}

	notif := &domain.Notification{
		LearnerID: payload.LearnerID,
		EventType: events.EventTypeConceptWeak,
		Payload:   marshalPayload(map[string]string{"message": "A weakness was detected in concept " + payload.ConceptID}),
		CreatedAt: time.Now(),
	}

	return h.repo.Create(ctx, notif)
}

func (h *EventHandler) HandleGoalAchieved(ctx context.Context, e events.Event) error {
	var payload events.GoalAchievedPayload
	if err := extractPayload(e.Payload, &payload); err != nil {
		return err
	}

	notif := &domain.Notification{
		LearnerID: payload.LearnerID,
		EventType: events.EventTypeGoalAchieved,
		Payload:   marshalPayload(map[string]string{"message": "Congratulations! You have achieved your goal " + payload.GoalID}),
		CreatedAt: time.Now(),
	}

	return h.repo.Create(ctx, notif)
}

func (h *EventHandler) HandleResourceFlagged(ctx context.Context, e events.Event) error {
	var payload events.ResourceFlaggedPayload
	if err := extractPayload(e.Payload, &payload); err != nil {
		return err
	}

	notif := &domain.Notification{
		LearnerID: payload.LearnerID, // Curators might need this too
		EventType: events.EventTypeResourceFlagged,
		Payload:   marshalPayload(map[string]string{"message": "Resource " + payload.ResourceID + " was flagged"}),
		CreatedAt: time.Now(),
	}

	return h.repo.Create(ctx, notif)
}

func extractPayload(raw any, dest any) error {
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func marshalPayload(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

func (h *EventHandler) Register(bus events.Bus) {
	bus.Subscribe(events.EventTypeCompetencyUpdated, h.HandleCompetencyUpdated)
	bus.Subscribe(events.EventTypeConceptWeak, h.HandleConceptWeak)
	bus.Subscribe(events.EventTypeGoalAchieved, h.HandleGoalAchieved)
	bus.Subscribe(events.EventTypeResourceFlagged, h.HandleResourceFlagged)
}
