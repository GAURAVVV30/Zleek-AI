package application

import (
	"strings"

	"github.com/hcl-backend/services/api-go/internal/diagnostics/domain"
)

// familiarityScale maps a selected option to a coverage percentage.
var familiarityScale = []struct {
	ID       string
	Label    string
	Coverage int
}{
	{"opt_familiar_1", "I haven't touched this topic yet", 10},
	{"opt_familiar_2", "I've seen it but can't apply it on my own", 30},
	{"opt_familiar_3", "I can work through it with guidance", 55},
	{"opt_familiar_4", "I can apply it independently", 75},
	{"opt_familiar_5", "I could teach this topic to others", 90},
}

// BuildQuestions renders one scalar readiness question per concept.
func BuildQuestions(concepts []domain.Concept) []domain.Question {
	questions := make([]domain.Question, 0, len(concepts))
	for i, c := range concepts {
		opts := make([]domain.QuestionOption, 0, len(familiarityScale))
		for _, s := range familiarityScale {
			opts = append(opts, domain.QuestionOption{ID: s.ID, Text: s.Label})
		}
		questions = append(questions, domain.Question{
			QuestionID:     c.NodeID,
			QuestionNumber: i + 1,
			TotalQuestions: len(concepts),
			ConceptID:      c.NodeID,
			ConceptName:    c.Name,
			Prompt:         "How comfortable are you with " + c.Name + "?",
			Options:        opts,
		})
	}
	return questions
}

// BuildQuestionsWithPrompts renders scalar readiness questions using custom dynamic prompt strings.
func BuildQuestionsWithPrompts(concepts []domain.Concept, prompts map[string]string) []domain.Question {
	questions := make([]domain.Question, 0, len(concepts))
	for i, c := range concepts {
		opts := make([]domain.QuestionOption, 0, len(familiarityScale))
		for _, s := range familiarityScale {
			opts = append(opts, domain.QuestionOption{ID: s.ID, Text: s.Label})
		}
		prompt := prompts[c.NodeID]
		if prompt == "" {
			prompt = "How comfortable are you with " + c.Name + "?"
		}
		questions = append(questions, domain.Question{
			QuestionID:     c.NodeID,
			QuestionNumber: i + 1,
			TotalQuestions: len(concepts),
			ConceptID:      c.NodeID,
			ConceptName:    c.Name,
			Prompt:         prompt,
			Options:        opts,
		})
	}
	return questions
}

func coverageFor(optionID string) int {
	for _, s := range familiarityScale {
		if s.ID == optionID {
			return s.Coverage
		}
	}
	return 10
}

func statusFor(coverage int) string {
	switch {
	case coverage >= 75:
		return "competent"
	case coverage >= 55:
		return "in_progress"
	default:
		return "gap"
	}
}

func computeResults(s *domain.Session) *domain.BaselineResults {
	rows := make([]domain.CoverageRow, 0, len(s.Concepts))
	total := 0
	for _, c := range s.Concepts {
		cleanID := strings.TrimSpace(c.NodeID)
		coverage := 0
		isFamiliarity := false
		for _, q := range s.Questions {
			if strings.TrimSpace(q.QuestionID) == cleanID && len(q.Options) > 0 && strings.Contains(q.Options[0].Text, "haven't touched") {
				isFamiliarity = true
				break
			}
		}

		userAnswer := strings.TrimSpace(s.Answers[cleanID])
		if userAnswer == "" {
			userAnswer = strings.TrimSpace(s.Answers[c.NodeID])
		}

		correctAnswer := strings.TrimSpace(s.CorrectAnswers[cleanID])
		if correctAnswer == "" {
			correctAnswer = strings.TrimSpace(s.CorrectAnswers[c.NodeID])
		}

		if !isFamiliarity && s.CorrectAnswers != nil && len(s.CorrectAnswers) > 0 {
			if userAnswer != "" && userAnswer == correctAnswer {
				coverage = 100
			}
		} else {
			coverage = coverageFor(userAnswer)
		}
		rows = append(rows, domain.CoverageRow{
			ConceptID:          cleanID,
			ConceptName:        c.Name,
			CoveragePercentage: coverage,
			Status:             statusFor(coverage),
		})
		total += coverage
	}
	overall := 0
	if len(rows) > 0 {
		overall = total / len(rows)
	}
	level := "Beginner"
	if overall >= 75 {
		level = "Advanced"
	} else if overall >= 55 {
		level = "Intermediate"
	}

	var gaps []string
	for _, r := range rows {
		if r.Status == "gap" {
			gaps = append(gaps, []string{r.ConceptName}...)
		}
	}
	if len(gaps) > 3 {
		gaps = gaps[:3]
	}
	if gaps == nil {
		gaps = []string{}
	}
	return &domain.BaselineResults{
		AssessedLevel:          level,
		OverallScorePercentage: overall,
		ConceptCoverage:        rows,
		TopGaps:                gaps,
	}
}
