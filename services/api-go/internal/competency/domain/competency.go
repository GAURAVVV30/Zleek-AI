package domain

import (
	"time"
)

type CompetencyState string

const (
	StateNotStarted   CompetencyState = "not_started"
	StateInProgress   CompetencyState = "in_progress"
	StateWeakEvidence CompetencyState = "weak_evidence"
	StateCompetent    CompetencyState = "competent"
)

type CompetencyRecord struct {
	LearnerID      string
	ConceptID      string
	State          CompetencyState
	LastEvidenceID *string
	UpdatedAt      time.Time
}

type CompetencyHistory struct {
	ID            string
	LearnerID     string
	ConceptID     string
	PreviousState *string
	NewState      string
	EvidenceID    string
	ChangedAt     time.Time
}

// Transition rule: competent is only achieved if previous was not_started, in_progress, weak_evidence.
func ValidTransition(current CompetencyState, next CompetencyState) bool {
	// A simple deterministic rule: you can always regress or progress based on the latest evidence.
	return true
}
