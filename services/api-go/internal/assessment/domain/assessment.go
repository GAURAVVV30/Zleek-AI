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
	ConceptID   string // public node id
	Type        AssessmentType
	Rubric      json.RawMessage
	Version     int
	GeneratedBy string // 'expert'|'ai'
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
	Score           float64
	Confidence      float64
	Result          string // 'competent', 'weak', 'inconclusive'
	Feedback        string
	RemediationHint string
}

// AnswerKeyOption is one multiple-choice option stored in the answer key.
type AnswerKeyOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// QuizAnswerKey is the persisted grading payload for a generated quiz item.
type QuizAnswerKey struct {
	Options []AnswerKeyOption `json:"options"`
	Correct string            `json:"correct"`
}

// QuizOption is a client-visible answer option.
type QuizOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// QuizQuestion is a client-visible generated question.
type QuizQuestion struct {
	ID      string       `json:"id"`
	Number  int          `json:"number"`
	Prompt  string       `json:"prompt"`
	Options []QuizOption `json:"options"`
}

// QuizView is the GET /concepts/{id}/assessment payload.
type QuizView struct {
	ConceptID    string         `json:"conceptId"`
	ConceptTitle string         `json:"conceptTitle"`
	Questions    []QuizQuestion `json:"questions"`
}

// AnswerSubmission is one answered question in the submit payload.
type AnswerSubmission struct {
	QuestionID       string `json:"questionId"`
	SelectedOptionID string `json:"selectedOptionId"`
}

// QuizSubmission is the POST /concepts/{id}/assessment/submit payload.
type QuizSubmission struct {
	Answers  []AnswerSubmission `json:"answers"`
	FreeText string             `json:"freeText,omitempty"`
}

// QuizResult is the submit response payload.
type QuizResult struct {
	Passed               bool    `json:"passed"`
	ScorePercentage      int     `json:"scorePercentage"`
	NewCompetencyState   string  `json:"newCompetencyState"`
	Feedback             string  `json:"feedback"`
	RemediationTriggered bool    `json:"remediationTriggered"`
	Score                float64 `json:"-"`
	Confidence           float64 `json:"-"`
	Result               string  `json:"-"`
}

// Evidence is the payload recorded through the progress pipeline after a
// submission is graded.
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
