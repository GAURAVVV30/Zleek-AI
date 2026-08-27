package domain

import (
	"encoding/json"
	"time"
)

type Evidence struct {
	ID                     string
	LearnerID              string
	ConceptID              string
	AssessmentDefinitionID string
	PathItemID             *string
	SubmissionData         json.RawMessage
	Score                  float64
	Confidence             float64
	EvaluatorType          string // 'ai'
	Result                 string
	CreatedAt              time.Time
}

type EngagementEvent struct {
	ID         string
	LearnerID  string
	PathItemID string
	EventType  string // e.g., 'view', 'interact'
	Timestamp  time.Time
}

type ProgressSummary struct {
	LearnerID       string
	CompetentCount  int
	InProgressCount int
	GoalsCompleted  int
}

type GoalCompletionSummary struct {
	GoalID        string
	TotalConcepts int
	Completed     int
}
