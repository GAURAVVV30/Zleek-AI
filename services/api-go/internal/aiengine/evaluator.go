package aiengine

// AssessmentEvaluator is a faithful Go port of app/core/evaluator.py.

import (
	"fmt"
	"math"
	"strings"
)

// SCORE_CORRECT_THRESHOLD matches the Python evaluator.
const scoreCorrectThreshold = 0.7

const evalBaseSystemPrompt = "" +
	"You are a strict pedagogical evaluator. Grade answers against the provided criteria." +
	" Provide a numeric score between 0.0 and 1.0, concise constructive feedback, and a remediation hint if the" +
	" answer does not meet the threshold. Be objective and base the score on the criteria only."

// EvaluateSubmission mirrors AssessmentEvaluator.evaluate_submission().
func (engine *GraphEngine) EvaluateSubmission(domainID, nodeID, studentAnswer string, attemptHistory []int, llm *LLMClient) map[string]any {
	// 1. Retrieve node and rubric.
	node, err := engine.GetNode(domainID, nodeID)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	rubric := node.AssessmentRubric
	diagnosticQuestion, _ := rubric["diagnostic_question"].(string)
	if diagnosticQuestion == "" {
		diagnosticQuestion = "Please evaluate the following response."
	}
	keyCriteria := anyStringSlice(rubric["key_evaluation_criteria"])
	remediationHintDefault, _ := rubric["remediation_hint"].(string)
	if remediationHintDefault == "" {
		remediationHintDefault = "Consider reviewing the core concepts and examples."
	}

	// Pull per-node BKT params from the graph and register them.
	if bpRaw, ok := rubric["bkt_params"].(map[string]any); ok {
		bp := DefaultBktParams()
		if v, ok := num(bpRaw["p_init"]); ok {
			bp.PInit = v
		}
		if v, ok := num(bpRaw["p_learn"]); ok {
			bp.PLearn = v
		}
		if v, ok := num(bpRaw["p_guess"]); ok {
			bp.PGuess = v
		}
		if v, ok := num(bpRaw["p_slip"]); ok {
			bp.PSlip = v
		}
		RegisterSkill(nodeID, bp)
	}

	// 2. Sentiment analysis BEFORE prompting the LLM.
	sentiment := AnalyzeSentiment(studentAnswer)
	toneOverride := sentiment.ToneOverride

	// 3. Build LLM prompts — tone-adjusted if frustration detected.
	systemPrompt := BuildToneSystemPrompt(evalBaseSystemPrompt, toneOverride)

	var criteriaText string
	if len(keyCriteria) > 0 {
		var b strings.Builder
		for i, c := range keyCriteria {
			fmt.Fprintf(&b, "%d. %s\n", i+1, c)
		}
		criteriaText = b.String()
	} else {
		criteriaText = "(no explicit criteria provided)"
	}

	userPrompt := fmt.Sprintf(
		"Diagnostic question: %s\nKey evaluation criteria:\n%s\n\nStudent answer:\n%s\n\n"+
			"Return ONLY valid JSON with the following fields: \n"+
			"- score: a number between 0.0 and 1.0\n"+
			"- passed: boolean (true if score >= %v)\n"+
			"- feedback: constructive, Socratic feedback guiding the student\n"+
			"- remediation_hint: (optional) short actionable hint if failed\n",
		diagnosticQuestion, criteriaText, studentAnswer, scoreCorrectThreshold,
	)

	schema := map[string]any{
		"score":            map[string]any{"type": "number", "min": 0.0, "max": 1.0},
		"passed":           map[string]any{"type": "boolean"},
		"feedback":         map[string]any{"type": "string"},
		"remediation_hint": map[string]any{"type": "string"},
	}

	// 4. LLM grading.
	result := llm.GenerateStructuredJSON(systemPrompt, userPrompt, schema)
	if e, ok := result["error"]; ok && e != nil {
		return map[string]any{"error": e, "raw": result["raw"]}
	}

	// Normalize score.
	scoreVal, ok := num(result["score"])
	if !ok {
		return map[string]any{"error": "Invalid or missing 'score' in LLM response.", "raw": result}
	}
	score := math.Max(0.0, math.Min(1.0, scoreVal))
	passed := score >= scoreCorrectThreshold

	feedback, _ := result["feedback"].(string)
	if feedback == "" {
		feedback = "No feedback provided."
	}
	remediation, _ := result["remediation_hint"].(string)
	if remediation == "" && !passed {
		remediation = remediationHintDefault
	}

	// 5. BKT mastery update.
	binarySignal := 0
	if passed {
		binarySignal = 1
	}
	history := append([]int(nil), attemptHistory...)
	history = append(history, binarySignal)
	bktResult := BktEstimate(nodeID, history, nil)

	// 6. Assemble enriched response.
	response := map[string]any{
		"score":         round3(score),
		"passed":        passed,
		"feedback":      feedback,
		"binary_signal": binarySignal,
		"bkt":           bktResult,
		"sentiment":     sentiment,
	}
	if !passed && remediation != "" {
		response["remediation_hint"] = remediation
	}
	return response
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

// tryEvaluateBetter handles the case where the LLM is unavailable and returns
// a deterministic rubric-based score, keeping the evaluate endpoint useful.
// Note: this is intentionally NOT used on the primary path so that BKT and
// mastery semantics stay identical to FastAPI.

func num(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case string:
		var f float64
		_, err := fmt.Sscanf(t, "%f", &f)
		return f, err == nil
	}
	return 0, false
}

func anyStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
