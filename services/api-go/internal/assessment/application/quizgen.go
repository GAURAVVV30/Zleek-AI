package application

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"

	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/keys"
)

const quizQuestionCount = 5

// generateQuiz builds a deterministic, 5-question multiple-choice assessment
// for a concept. The same concept id always yields the same quiz, so grade
// keys and evidence stay reproducible without an LLM.
func generateQuiz(conceptID, conceptTitle string, core []string) (*domain.AssessmentDefinition, []domain.AssessmentItem) {
	topics := nonEmpty(core, []string{"the fundamentals", "prerequisites", "hands-on practice", "foundational skills"})
	k0, k1, k2 := topics[0], pick(topics, 1, topics[0]), pick(topics, 2, topics[1])

	questions := []QuizQuestionSpec{
		{
			prompt:  fmt.Sprintf("Which statement best describes the purpose of %s?", conceptTitle),
			correct: fmt.Sprintf("It focuses on building practical %s skills through structured study and practice.", conceptTitle),
			distractors: []string{
				"It removes the need for any prerequisite study.",
				"It covers only the theoretical history of the field.",
				fmt.Sprintf("It is unrelated to %s.", k0),
			},
		},
		{
			prompt:  fmt.Sprintf("What is the most important foundation for %s?", conceptTitle),
			correct: fmt.Sprintf("%s, because %s builds directly on it.", k0, conceptTitle),
			distractors: []string{
				"Advanced topics unrelated to " + conceptTitle + ".",
				"No foundation is required at all.",
				fmt.Sprintf("%s only, skipping %s.", k1, k0),
			},
		},
		{
			prompt:  fmt.Sprintf("Which order should you follow before attempting %s?", conceptTitle),
			correct: fmt.Sprintf("%s, then %s.", k0, k1),
			distractors: []string{
				conceptTitle + " itself before anything else.",
				"An unrelated " + k2 + " area.",
				"None; " + conceptTitle + " is entry-level.",
			},
		},
		{
			prompt:  fmt.Sprintf("What role does %s play in mastering %s?", k0, conceptTitle),
			correct: fmt.Sprintf("It provides the core concepts %s relies on, such as %s.", conceptTitle, k1),
			distractors: []string{
				"It is unrelated to practical work.",
				"It is only a naming convention.",
				fmt.Sprintf("It replaces the need for %s.", k2),
			},
		},
		{
			prompt:  fmt.Sprintf("Which statement about the scope of %s is correct?", conceptTitle),
			correct: "It covers selected sub-topics that must be practiced incrementally.",
			distractors: []string{
				"It covers every known topic instantly.",
				"It has no recommended learning order.",
				"It cannot be applied to real projects.",
			},
		},
	}

	def := &domain.AssessmentDefinition{
		ID:          deterministicKey(conceptID + ":def"),
		ConceptID:   conceptID,
		Type:        domain.TypeQuiz,
		Rubric:      mustJSON(map[string]any{"type": "quiz", "questions": quizQuestionCount, "threshold": 0.7, "generator": "deterministic"}),
		Version:     1,
		GeneratedBy: "expert",
	}

	items := make([]domain.AssessmentItem, 0, quizQuestionCount)
	for i, q := range questions {
		options := []domain.AnswerKeyOption{
			{ID: fmt.Sprintf("opt_%d_a", i+1), Text: q.correct},
			{ID: fmt.Sprintf("opt_%d_b", i+1), Text: q.distractors[0]},
			{ID: fmt.Sprintf("opt_%d_c", i+1), Text: q.distractors[1]},
			{ID: fmt.Sprintf("opt_%d_d", i+1), Text: q.distractors[2]},
		}
		// Deterministically shuffle so the correct option's position varies.
		r := rand.New(rand.NewSource(int64(seedFor(conceptID, i))))
		r.Shuffle(len(options), func(a, b int) { options[a], options[b] = options[b], options[a] })
		key := domain.QuizAnswerKey{Options: options, Correct: options[0].ID}
		items = append(items, domain.AssessmentItem{
			ID:                     deterministicKey(fmt.Sprintf("%s:item:%d", conceptID, i+1)),
			AssessmentDefinitionID: def.ID,
			Prompt:                 q.prompt,
			ItemType:               "multiple_choice",
			AnswerKey:              mustJSON(key),
		})
	}
	return def, items
}

type QuizQuestionSpec struct {
	prompt      string
	correct     string
	distractors []string
}

func pick(s []string, i int, fallback string) string {
	if len(s) > i {
		return s[i]
	}
	if len(s) > 0 {
		return s[len(s)-1]
	}
	return fallback
}

func nonEmpty(s, fallback []string) []string {
	if len(s) > 0 {
		return s
	}
	return fallback
}

func seedFor(conceptID string, i int) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s:%d", conceptID, i)))
	return h.Sum64()
}

func deterministicKey(s string) string {
	return keys.UUID("quiz:" + s)
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

// questionsForView converts items to the client-visible quiz shape.
func questionsForView(items []domain.AssessmentItem) []domain.QuizQuestion {
	out := make([]domain.QuizQuestion, 0, len(items))
	for i, item := range items {
		var key domain.QuizAnswerKey
		if json.Unmarshal(item.AnswerKey, &key) == nil {
			opts := make([]domain.QuizOption, 0, len(key.Options))
			for _, o := range key.Options {
				opts = append(opts, domain.QuizOption{ID: o.ID, Text: o.Text})
			}
			out = append(out, domain.QuizQuestion{
				ID:      item.ID,
				Number:  i + 1,
				Prompt:  item.Prompt,
				Options: opts,
			})
			continue
		}
		out = append(out, domain.QuizQuestion{
			ID:     item.ID,
			Number: i + 1,
			Prompt: item.Prompt,
			Options: []domain.QuizOption{
				{ID: "opt_a", Text: "Option A"},
				{ID: "opt_b", Text: "Option B"},
				{ID: "opt_c", Text: "Option C"},
			},
		})
	}
	return out
}
