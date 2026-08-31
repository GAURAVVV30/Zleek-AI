package aiengine

// GuardrailsEngine is a Go port of app/core/guardrails_engine.py.
// NeMo Guardrails is Python-only; the deterministic keyword guardrail —
// identical patterns, reasons, and refusal messages — is the runtime path.

import (
	"os"
	"regexp"
	"strings"
)

// blockPattern mirrors the _BLOCK_PATTERNS lists.
var blockPatterns = []struct {
	name    string
	regex   []*regexp.Regexp
	refusal string
}{
	{
		name: "direct_answer_request",
		regex: compileNFAS(
			`\bjust\s+(give|tell|show)\s+me\s+the\s+(answer|solution|result)\b`,
			`\bwhat\s+(is|are)\s+the\s+(correct\s+)?answer\b`,
			`\btell\s+me\s+the\s+solution\b`,
			`\bwhat.{0,10}answer\b`,
		),
		refusal: "I'm your Socratic tutor — my role is to guide you to the answer, not give it to you. " +
			"Let me ask: what do you already know about this concept? Start there and we'll build up together.",
	},
	{
		name: "code_generation_request",
		regex: compileNFAS(
			`\b(write|generate|give me|create|make|produce)\s+.{0,20}(code|script|function|program|solution)\s*(for me|please|now)?\b`,
			`\bjust\s+write\s+(it|the code|the solution)\b`,
			`\bwrite\s+(a\s+)?(python|java|javascript|sql|c\+\+|go|rust)\s+(code|script|function|program)\b`,
			`\bcomplete\s+(the\s+)?(code|function|program)\s*(for me)?\b`,
			`\bfinish\s+the\s+(code|function)\b`,
			`\bgive\s+me\s+(the\s+)?(code|solution|answer)\b`,
			`\bcode\s+(for|to)\s+me\b`,
		),
		refusal: "I'm here to help you become a developer, not to be your code generator. " +
			"Let's break this down: what's the first logical step you'd take? " +
			"Think about the inputs and outputs first — what should the function receive and return?",
	},
	{
		name: "homework_completion_request",
		regex: compileNFAS(
			`\bdo\s+(my\s+)?(assignment|homework|task|project)\b`,
			`\bcomplete\s+(this\s+)?(for me|assignment|task)\b`,
			`\bsolve\s+this\s+for me\b`,
			`\bwrite\s+my\s+(essay|report|assignment)\b`,
		),
		refusal: "Completing assignments for you would rob you of the learning experience. " +
			"Instead, let's tackle this together. " +
			"What part of the problem statement is unclear? Let's start with the simplest piece.",
	},
}

// compileNFAS builds regexes tolerantly so a bad pattern can't crash startup.
func compileNFAS(patterns ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		out = append(out, re)
	}
	return out
}

// GuardrailsResult mirrors the dict returned by GuardrailsEngine.check().
type GuardrailsResult struct {
	Blocked        bool   `json:"blocked"`
	Reason         string `json:"reason"`
	RefusalMessage string `json:"refusal_message"`
	Method         string `json:"method"`
	OriginalText   string `json:"original_text"`
}

// keywordCheck mirrors _keyword_check().
func keywordCheck(text string) *GuardrailsResult {
	lower := strings.ToLower(strings.TrimSpace(text))
	for _, bp := range blockPatterns {
		for _, re := range bp.regex {
			if re.MatchString(lower) {
				return &GuardrailsResult{
					Blocked:        true,
					Reason:         bp.name,
					RefusalMessage: bp.refusal,
					Method:         "keyword_guardrail",
					OriginalText:   text,
				}
			}
		}
	}
	return nil
}

// CheckGuardrails mirrors GuardrailsEngine.check() (keyword path; NeMo is not
// ported to Go so the semantic rail is reported as unavailable).
func CheckGuardrails(studentText string) GuardrailsResult {
	if strings.TrimSpace(studentText) == "" {
		return GuardrailsResult{Blocked: false, Method: "passthrough", OriginalText: studentText}
	}

	if keywordResult := keywordCheck(studentText); keywordResult != nil {
		return *keywordResult
	}

	return GuardrailsResult{
		Blocked:      false,
		Method:       "keyword_guardrail",
		OriginalText: studentText,
	}
}

// GuardrailsStatus mirrors status(): NeMo never loads in the Go port.
func GuardrailsStatus() map[string]any {
	return map[string]any{
		"nemo_available":     false,
		"keyword_fallback":   true,
		"config_dir":         "data/guardrails (not ported; NeMo is Python-only)",
		"nvidia_key_present": strings.TrimSpace(os.Getenv("NVIDIA_API_KEY")) != "",
		"method":             "keyword_guardrail",
	}
}
