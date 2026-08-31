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
	EvaluatorType          string // 'ai'|'curator'
	Result                 string // 'competent'|'weak'|'inconclusive'
	CreatedAt              time.Time
}

type EngagementEvent struct {
	ID         string
	LearnerID  string
	ConceptID  string
	PathItemID string
	EventType  string // 'resource_opened'|'marked_reviewed'
	Timestamp  time.Time
}

// SummaryRow is one competency bar from the progress dashboard.
type SummaryRow struct {
	Domain     string `json:"domain"`
	Percentage int    `json:"percentage"`
	Status     string `json:"status"`
}

// Summary is the /progress/summary payload.
type Summary struct {
	OverallCompletionPercentage int          `json:"overallCompletionPercentage"`
	TotalConcepts               int          `json:"totalConcepts"`
	CompletedConcepts           int          `json:"completedConcepts"`
	ActiveRemediations          int          `json:"activeRemediations"`
	CompetencyBreakdown         []SummaryRow `json:"competencyBreakdown"`
}

type GoalCompletionSummary struct {
	GoalID              string   `json:"goalId"`
	GoalTitle           string   `json:"goalTitle"`
	CompletionDate      string   `json:"completionDate"`
	TotalSkillsVerified int      `json:"totalSkillsVerified"`
	MasteryProofList    []string `json:"masteryProofList"`
}
