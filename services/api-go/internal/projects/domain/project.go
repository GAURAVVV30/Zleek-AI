package domain

import (
	"encoding/json"
	"time"
)

type ProjectStatus string

const (
	ProjectStatusSubmitted ProjectStatus = "submitted"
	ProjectStatusPending   ProjectStatus = "pending"
	ProjectStatusReviewed  ProjectStatus = "reviewed"
)

type Project struct {
	ID                     string
	ConceptID              string
	AssessmentDefinitionID string
	Rubric                 json.RawMessage
	Prompts                []string
}

type SubmissionMetadata struct {
	ArtifactReference string `json:"artifact_reference"`
	Notes             string `json:"notes,omitempty"`
}

type ProjectSubmission struct {
	LearnerID   string
	ConceptID   string
	Metadata    SubmissionMetadata
	SubmittedAt time.Time
}

type ProjectState struct {
	Status    ProjectStatus
	Result    string
	Score     *float64
	UpdatedAt time.Time
}
