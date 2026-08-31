package aiengine

// SentimentAnalyzer is a Go port of app/core/sentiment_analyzer.py.
// The HuggingFace transformer pipeline is Python-only; the keyword heuristic
// fallback is the runtime path here, preserving identical labels/confidence.

import (
	"regexp"
	"strings"
)

// Emotion -> tone mapping (Python FRUSTRATION_EMOTIONS / POSITIVE_EMOTIONS).
var frustrationEmotions = map[string]bool{"anger": true, "disgust": true, "fear": true}

// keywordPatterns mirrors _KEYWORD_PATTERNS.from the Python module.
var keywordPatterns = map[string][]*regexp.Regexp{
	"anger": compileAll(
		`\bstupid\b`, `\bdumb\b`, `\bconfusing\b`, `\bconfused\b`,
		`\bfrustrat\w*\b`, `\bawful\b`, `\bwrong\b`, `\bhate\b`,
		`\bcan'?t\b.{0,20}\bunderstand\b`, `\bmakes no sense\b`,
	),
	"sadness": compileAll(
		`\bgive up\b`, `\bcan'?t do\b`, `\bimpossible\b`, `\bgave up\b`,
		`\bdisappoin\w*\b`, `\btoo hard\b`,
	),
	"fear": compileAll(
		`\bscared\b`, `\bnervous\b`, `\banxious\b`, `\bworried\b`,
		`\bnot sure\b`, `\bnot confident\b`,
	),
	"joy": compileAll(
		`\bgot it\b`, `\bunderstand\b`, `\bclear\b`, `\beasy\b`,
		`\bmakes sense\b`, `\bgreat\b`, `\bexcit\w*\b`,
	),
}

func compileAll(patterns ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		out = append(out, regexp.MustCompile(p))
	}
	return out
}

// SentimentResult mirrors the dict returned by SentimentAnalyzer.analyze().
type SentimentResult struct {
	Emotion      string  `json:"emotion"`
	Confidence   float64 `json:"confidence"`
	ToneOverride string  `json:"tone_override"`
	Method       string  `json:"method"`
	RawLabel     string  `json:"raw_label"`
}

// AnalyzeSentiment mirrors SentimentAnalyzer.analyze() (heuristic path).
func AnalyzeSentiment(text string) SentimentResult {
	if strings.TrimSpace(text) == "" {
		return SentimentResult{
			Emotion:      "neutral",
			Confidence:   1.0,
			ToneOverride: "standard",
			Method:       "empty_input",
			RawLabel:     "neutral",
		}
	}

	lower := strings.ToLower(text)
	emotion := "neutral"
	confidence := 0.90
	// CHECK order: iterate the fixed dict order (anger, sadness, fear, joy, neutral).
	for _, key := range []string{"anger", "sadness", "fear", "joy"} {
		for _, re := range keywordPatterns[key] {
			if re.MatchString(lower) {
				emotion = key
				confidence = 0.70
				break
			}
		}
		if emotion != "neutral" {
			break
		}
	}

	tone := "standard"
	if frustrationEmotions[emotion] {
		tone = "encouraging"
	}

	return SentimentResult{
		Emotion:      emotion,
		Confidence:   confidence,
		ToneOverride: tone,
		Method:       "keyword_heuristic",
		RawLabel:     emotion,
	}
}

// BuildToneSystemPrompt mirrors build_tone_system_prompt().
func BuildToneSystemPrompt(baseSystemPrompt, toneOverride string) string {
	if toneOverride != "encouraging" {
		return baseSystemPrompt
	}
	tonePrefix := "IMPORTANT: The learner appears frustrated or confused. " +
		"Your response must be exceptionally patient, warm, and encouraging. " +
		"Break down the concept into the smallest possible steps. " +
		"Celebrate any partial understanding. " +
		"Avoid technical jargon — use relatable analogies instead. " +
		"Start your feedback with a genuine, uplifting acknowledgement of their effort.\n\n"
	return tonePrefix + baseSystemPrompt
}
