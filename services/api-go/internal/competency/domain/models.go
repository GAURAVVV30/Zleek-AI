package domain

import (
	"time"
)

type State string

const (
	StateNotStarted State = "not_started"
	StateInProgress State = "in_progress"
	StateWeak       State = "weak_evidence"
	StateCompetent  State = "competent"
)

// CompetencyRecord is the learner's current competency for a concept. IDs are
// the public roadmap node ids.
type CompetencyRecord struct {
	ConceptID        string    `json:"conceptId"`
	ConceptName      string    `json:"conceptName"`
	State            string    `json:"state"`
	Score            float64   `json:"score"`
	LastEvidenceDate string    `json:"lastEvidenceDate"`
	EvidenceType     string    `json:"evidenceType"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// CompetencyHistory is one transition record per evidence event.
type CompetencyHistory struct {
	Attempt       int       `json:"attempt"`
	Date          string    `json:"date"`
	Score         float64   `json:"score"`
	Result        string    `json:"result"`
	Details       string    `json:"details"`
	PreviousState string    `json:"previousState"`
	NewState      string    `json:"newState"`
	ChangedAt     time.Time `json:"changedAt"`
}
