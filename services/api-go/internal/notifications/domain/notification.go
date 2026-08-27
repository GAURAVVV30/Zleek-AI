package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrUnauthorized         = errors.New("unauthorized")
)

type Notification struct {
	ID        string          `json:"id"`
	LearnerID string          `json:"learner_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	ReadAt    *time.Time      `json:"read_at"`
	CreatedAt time.Time       `json:"created_at"`
}

func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}
