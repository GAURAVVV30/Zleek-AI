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
	ConceptID     string // public roadmap node id
	ResourceID    *string
	SequenceOrder int
	State         PathItemState
	IsRemediation bool
	InsertedAt    time.Time
}

type RemediationRecord struct {
	ID                    string
	LearnerID             string
	ConceptID             string
	TriggeredByEvidenceID string
	RemediationResourceID *string
	AttemptNumber         int
	Status                RemediationStatus
	CreatedAt             time.Time
	ResolvedAt            *time.Time
}

// RoadmapNode is the client-visible roadmap milestone shape.
type RoadmapNode struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Domain            string `json:"domain"`
	DomainID          string `json:"domain_id,omitempty"`
	State             string `json:"state"`
	Order             int    `json:"order"`
	EstimatedMinutes  int    `json:"estimatedMinutes"`
	IsRemediation     bool   `json:"isRemediation"`
	UnlockRequirement string `json:"unlockRequirement,omitempty"`
	NextSubConcept    string `json:"nextSubConcept,omitempty"`
}

// Roadmap is the GET /roadmap payload.
type Roadmap struct {
	GoalID             string        `json:"goalId"`
	GoalTitle          string        `json:"goalTitle"`
	Domain             string        `json:"domain"`
	DomainID           string        `json:"domain_id"`
	ProgressPercentage int           `json:"progressPercentage"`
	CurrentNodeID      string        `json:"currentNodeId"`
	Nodes              []RoadmapNode `json:"nodes"`
}

type DailyTask struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Duration  int    `json:"duration"`
	Completed bool   `json:"completed"`
}

type DailyTaskDay struct {
	Date  time.Time   `json:"date"`
	Tasks []DailyTask `json:"tasks"`
}
