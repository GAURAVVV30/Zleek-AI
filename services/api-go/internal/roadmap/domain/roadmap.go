package domain

import (
	"time"
)

type PathStatus string

const (
	PathStatusActive    PathStatus = "active"
	PathStatusCompleted PathStatus = "completed"
)

type PathItemState string

const (
	ItemStateLocked       PathItemState = "locked"
	ItemStateAvailable    PathItemState = "available"
	ItemStateInProgress   PathItemState = "in_progress"
	ItemStateWeakEvidence PathItemState = "weak_evidence"
	ItemStateCompetent    PathItemState = "competent"
)

type RemediationStatus string

const (
	RemediationStatusPending   RemediationStatus = "pending"
	RemediationStatusResolved  RemediationStatus = "resolved"
	RemediationStatusEscalated RemediationStatus = "escalated"
)

type Path struct {
	ID                   string
	LearnerID            string
	GoalID               string
	KnowledgeStructureID string
	Status               PathStatus
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type PathItem struct {
	ID            string
	PathID        string
	ConceptID     string
	ResourceID    *string
	SequenceOrder int
	State         PathItemState
	IsRemediation bool
	InsertedAt    time.Time
}

type RemediationRecord struct {
	ID                     string
	LearnerID              string
	ConceptID              string
	TriggeredByEvidenceID  string
	RemediationResourceID  *string
	AttemptNumber          int
	Status                 RemediationStatus
	CreatedAt              time.Time
	ResolvedAt             *time.Time
}
