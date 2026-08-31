package aiengine

import (
	"testing"
)

func getTestEngine(t *testing.T) *GraphEngine {
	t.Helper()
	engine, err := ParseGraphEngine()
	if err != nil {
		t.Fatalf("ParseGraphEngine() error: %v", err)
	}
	return engine
}

func TestGraphEngineLoadsSeededDomains(t *testing.T) {
	engine := getTestEngine(t)
	if len(engine.DomainList) != 12 {
		t.Fatalf("expected 12 seeded domains, got %d", len(engine.DomainList))
	}
	for _, want := range []string{"machine_learning", "software_architecture", "data_engineer", "frontend_engineer"} {
		if !engine.DomainExists(want) {
			t.Errorf("expected domain %q to be loaded", want)
		}
	}
	if _, err := engine.GetNode("machine_learning", "definitely_not_a_node"); err == nil {
		t.Error("expected error for unknown node")
	}
}

func TestPersonalizedPathRespectsPrerequisites(t *testing.T) {
	engine := getTestEngine(t)
	path, err := engine.GetPersonalizedPath("machine_learning", nil)
	if err != nil {
		t.Fatalf("GetPersonalizedPath error: %v", err)
	}
	if len(path) == 0 {
		t.Fatal("expected a non-empty learning path")
	}
	positions := map[string]int{}
	for i, node := range path {
		id, _ := node["id"].(string)
		positions[id] = i
	}
	for _, node := range path {
		id, _ := node["id"].(string)
		prereqs, _ := node["prerequisites"].(map[string]any)
		hard, _ := prereqs["hard"].([]any)
		for _, dep := range hard {
			depStr, _ := dep.(string)
			if pos, ok := positions[depStr]; ok && pos >= positions[id] {
				t.Errorf("node %q appears before prerequisite %q", id, depStr)
			}
		}
	}
}

func TestPersonalizedPathExcludesCompleted(t *testing.T) {
	engine := getTestEngine(t)
	full, _ := engine.GetPersonalizedPath("software_architecture", nil)
	if len(full) == 0 {
		t.Fatal("expected non-empty path")
	}
	firstID, _ := full[0]["id"].(string)
	remaining, err := engine.GetPersonalizedPath("software_architecture", []string{firstID})
	if err != nil {
		t.Fatalf("GetPersonalizedPath error: %v", err)
	}
	if len(remaining) != len(full)-1 {
		t.Fatalf("expected %d remaining nodes, got %d", len(full)-1, len(remaining))
	}
}

func TestPersonalizedPathUnknownDomain(t *testing.T) {
	engine := getTestEngine(t)
	if _, err := engine.GetPersonalizedPath("nope", nil); err == nil {
		t.Error("expected error for unknown domain")
	}
}

func TestBKTEstimate(t *testing.T) {
	res := BktEstimate("test.node", []int{1, 1, 1}, nil)
	if res.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", res.Attempts)
	}
	if res.Params["p_init"] != 0.30 || res.Params["p_learn"] != 0.20 || res.Params["p_guess"] != 0.25 || res.Params["p_slip"] != 0.10 {
		t.Errorf("unexpected default params: %v", res.Params)
	}
	if res.PMastery < 0.9 {
		t.Errorf("expected high mastery after 3 correct, got %v", res.PMastery)
	}
	if !res.Mastered {
		t.Error("expected mastered after 3 correct with default threshold")
	}
	res2 := BktEstimate("test.node", []int{1, 1, 1, 1, 1}, nil)
	if !res2.Mastered {
		t.Error("expected mastered after 5 consecutive correct")
	}
}

func TestBKTCustomParams(t *testing.T) {
	cp := DefaultBktParams()
	cp.PInit = 0.9 // learner almost certainly knows it
	res := BktEstimate("custom.node", []int{1}, &cp)
	if res.PMastery < 0.9 {
		t.Errorf("expected near-initial mastery, got %v", res.PMastery)
	}
}

func TestBKTIncremental(t *testing.T) {
	res := BktEstimateIncremental("inc.node", 0.5, 0, nil)
	if res.Attempts != 1 {
		t.Errorf("expected single attempt, got %d", res.Attempts)
	}
	if res.PMastery >= 0.99 {
		t.Error("wrong answer should not push mastery to ceiling")
	}
}

func TestGuardrailsBlocksRequests(t *testing.T) {
	cases := map[string]string{
		"Just write the python code for me": "code_generation_request",
		"Tell me the solution right now":    "direct_answer_request",
		"Complete this task for me":         "homework_completion_request",
	}
	for input, wantReason := range cases {
		res := CheckGuardrails(input)
		if !res.Blocked || res.Reason != wantReason {
			t.Errorf("input %q -> blocked=%v reason=%q, want reason %q", input, res.Blocked, res.Reason, wantReason)
		}
	}
}

func TestGuardrailsAllowsSocratic(t *testing.T) {
	res := CheckGuardrails("I'm struggling to understand how to structure the code, can you help me think it through?")
	if res.Blocked {
		t.Errorf("expected allowed, got %+v", res)
	}
	if res.Method != "keyword_guardrail" {
		t.Errorf("expected keyword_guardrail method, got %s", res.Method)
	}
}

func TestSentiment(t *testing.T) {
	res := AnalyzeSentiment("This is so confusing and frustrating")
	if res.Emotion != "anger" {
		t.Errorf("expected anger, got %s", res.Emotion)
	}
	if res.ToneOverride != "encouraging" {
		t.Errorf("expected encouraging tone, got %s", res.ToneOverride)
	}
	neutral := AnalyzeSentiment("  ")
	if neutral.Emotion != "neutral" || neutral.Method != "empty_input" {
		t.Errorf("expected empty-input neutral, got %+v", neutral)
	}
}

func TestAdaptiveNextAction(t *testing.T) {
	pm := 0.96
	advance := DetermineNextAction("node", 0.5, 0, &pm)
	if advance.Action != "advance" || advance.DecisionBasis != "bkt_mastery" {
		t.Errorf("expected bkt advance, got %+v", advance)
	}

	remediate := DetermineNextAction("node", 0.5, 1, nil)
	if remediate.Action != "remediate" {
		t.Errorf("expected remediate, got %+v", remediate)
	}

	intervene := DetermineNextAction("node", 0.5, 3, nil)
	if intervene.Action != "human_intervention" {
		t.Errorf("expected human_intervention, got %+v", intervene)
	}

	legacy := DetermineNextAction("node", 0.85, 0, nil)
	if legacy.Action != "advance" || legacy.DecisionBasis != "legacy_score" {
		t.Errorf("expected legacy score advance, got %+v", legacy)
	}
}

func TestMatchDomainByKeywords(t *testing.T) {
	engine := getTestEngine(t)
	cases := map[string]string{
		"I want to become a machine learning engineer": "machine_learning",
		"I am aiming for a data analyst role":          "data_analyst",
		"Want to build backend apis":                   "backend_engineer",
	}
	for input, want := range cases {
		if got := engine.MatchDomainByKeywords(input); got != want {
			t.Errorf("MatchDomainByKeywords(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestEvaluateSubmissionWithoutLLM(t *testing.T) {
	engine := getTestEngine(t)
	llm := DefaultLLM() // no API keys in test env
	res := engine.EvaluateSubmission("machine_learning", "ml_01_programming_fundamentals", "I understand python", nil, llm)
	if e, ok := res["error"]; ok {
		// Without keys, the pipeline degrades to a clean error dict like Python.
		_ = e
		return
	}
	if _, ok := res["score"]; !ok {
		t.Errorf("expected score in response, got %+v", res)
	}
}

func TestLLMStructuredJSONDegrades(t *testing.T) {
	llm := DefaultLLM()
	res := llm.GenerateStructuredJSON("sys", "prompt", map[string]any{"a": map[string]any{"type": "string"}})
	if _, ok := res["error"]; !ok {
		t.Errorf("expected error dict without API keys, got %+v", res)
	}
}

func TestRoadmapStoreRoundtrip(t *testing.T) {
	rs := NewRoadmapStore()
	if got := rs.Get("ml_engineer"); got["domain"] == nil {
		t.Error("expected seeded ml_engineer roadmap")
	}
	record := BuildRoadmapRecord("test domain", []map[string]any{{"id": "a", "label": "A"}}, []map[string]any{{"source": "a", "target": "b"}})
	if err := rs.Put("test domain", record); err != nil {
		t.Fatalf("Put error: %v", err)
	}
	got := rs.Get("test domain")
	if got == nil || got["domain"] != "test domain" {
		t.Errorf("expected persisted record, got %+v", got)
	}
	found := false
	for _, d := range rs.List() {
		if d == "test_domain" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected test_domain in List(), got %v", rs.List())
	}
}

func TestGetResourcesForNode(t *testing.T) {
	engine := getTestEngine(t)
	res := engine.GetResourcesForNode("", "ml_01_programming_fundamentals", 3)
	if len(res) == 0 {
		t.Fatal("expected resources for python_basics")
	}
	if res[0].Title == "" {
		t.Error("expected resource title")
	}
	if res[0].ID == "" {
		t.Error("expected synthesized resource id")
	}
}

func TestVoiceStatusDegrades(t *testing.T) {
	v := NewVoiceEngine() // no keys
	s := v.Status()
	if s["asr_available"] != false {
		t.Errorf("expected asr unavailable, got %v", s["asr_available"])
	}
	if _, err := v.Synthesize("hello"); err == nil {
		t.Error("expected synthesize failure without NVIDIA key")
	}
	if _, _, err := v.Transcribe([]byte("x"), "audio/wav"); err == nil {
		t.Error("expected transcribe failure without ASR keys")
	}
}

func TestAnalyzeUserIntentDegrades(t *testing.T) {
	engine := getTestEngine(t)
	res := engine.AnalyzeUserIntent("become a machine learning engineer", DefaultLLM())
	if e, ok := res["error"]; ok {
		t.Logf("degraded as expected: %v", e)
	} else if id, _ := res["mapped_domain_id"].(string); id == "" {
		t.Errorf("expected mapped domain id")
	}
}
