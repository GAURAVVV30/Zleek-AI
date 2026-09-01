package domain

// QuestionOption is one answer choice of a diagnostic question.
type QuestionOption struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect,omitempty"`
}

// Question is one diagnostic question payload.
type Question struct {
	QuestionID     string           `json:"questionId"`
	QuestionNumber int              `json:"questionNumber"`
	TotalQuestions int              `json:"totalQuestions"`
	ConceptID      string           `json:"conceptId"`
	ConceptName    string           `json:"conceptName"`
	Prompt         string           `json:"prompt"`
	Options        []QuestionOption `json:"options"`
}

// Session models an in-progress diagnostic for a learner.
type Session struct {
	SessionID      string
	LearnerID      string
	Concepts       []Concept
	Answers        map[string]string // questionID -> optionID
	Prompts        map[string]string // questionID -> dynamic AI prompt text
	CorrectAnswers map[string]string
	Questions      []Question
	Completed      bool
}

// Concept is a graded concept within a session.
type Concept struct {
	NodeID string
	Name   string
}

// AnswerResponse is the per-answer payload.
type AnswerResponse struct {
	IsComplete       bool      `json:"isComplete"`
	IsCorrect        bool      `json:"isCorrect"`
	CorrectOptionID  string    `json:"correctOptionId"`
	SelectedOptionID string    `json:"selectedOptionId"`
	NextQuestion     *Question `json:"nextQuestion,omitempty"`
}

// CoverageRow is one concept row of the baseline results.
type CoverageRow struct {
	ConceptID          string `json:"conceptId"`
	ConceptName        string `json:"conceptName"`
	CoveragePercentage int    `json:"coveragePercentage"`
	Status             string `json:"status"` // 'competent'|'in_progress'|'gap'
}

// BaselineResults is the GET /diagnostics/{sessionId}/results payload.
type BaselineResults struct {
	AssessedLevel          string        `json:"assessedLevel"`
	OverallScorePercentage int           `json:"overallScorePercentage"`
	ConceptCoverage        []CoverageRow `json:"conceptCoverage"`
	TopGaps                []string      `json:"topGaps"`
	Explanation            string        `json:"explanation"`
}
