package domain

import (
	"encoding/json"
	"time"
)

type AssessmentType string

const (
	TypeQuiz     AssessmentType = "quiz"
	TypeScenario AssessmentType = "scenario"
	TypeProject  AssessmentType = "project"
)

type AssessmentDefinition struct {
	ID          string
	ConceptID   string
	Type        AssessmentType
	Rubric      json.RawMessage
	Version     int
	GeneratedBy string // 'expert' or 'ai'
	CreatedAt   time.Time
	Items       []AssessmentItem
}

type AssessmentItem struct {
	ID                     string
	AssessmentDefinitionID string
	Prompt                 string
	ItemType               string
	AnswerKey              json.RawMessage
	CreatedAt              time.Time
}

type EvaluationResult struct {
	Score      float64
	Confidence float64
	Result     string // 'competent', 'weak', 'inconclusive'
}

type Evidence struct {
	ID                     string
	LearnerID              string
	ConceptID              string
	AssessmentDefinitionID string
	SubmissionData         json.RawMessage
	Score                  float64
	Confidence             float64
	EvaluatorType          string // 'ai'
	Result                 string
}

func (ad *AssessmentDefinition) IsValidType() bool {
	return ad.Type == TypeQuiz || ad.Type == TypeScenario || ad.Type == TypeProject
}
