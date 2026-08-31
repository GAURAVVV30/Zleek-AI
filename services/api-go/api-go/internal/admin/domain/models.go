package domain

import (
	"encoding/json"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditRecord struct {
	ID               string          `json:"id"`
	ActorID          string          `json:"actor_id"`
	Action           string          `json:"action"`
	TargetEntityType string          `json:"target_entity_type"`
	TargetEntityID   string          `json:"target_entity_id"`
	BeforeState      json.RawMessage `json:"before_state,omitempty"`
	AfterState       json.RawMessage `json:"after_state,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}
