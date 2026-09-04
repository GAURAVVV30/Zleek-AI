package aihttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error: %v", err)
	}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func doJSON(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var rd *bytes.Reader
	if body == "" {
		rd = bytes.NewReader(nil)
	} else {
		rd = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestRootAndHealth(t *testing.T) {
	h := newTestRouter(t)
	if rr := doJSON(t, h, http.MethodGet, "/", ""); rr.Code != http.StatusOK {
		t.Fatalf("root status %d", rr.Code)
	}
	rr := doJSON(t, h, http.MethodGet, "/api/v1/health", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("health status %d", rr.Code)
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["service"] != "ai-fastapi" || out["version"] != "3.0" {
		t.Errorf("unexpected health payload: %s", rr.Body.String())
	}
}

func TestGoalAnalyze(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/goal/analyze", `{"user_text":"I want to do machine learning"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	// With no LLM keys this may degrade to an error dict; with keys it maps.
	if _, hasErr := out["error"]; hasErr {
		t.Logf("degraded goal analyze: %v", out["error"])
	} else if out["mapped_domain_id"] == nil {
		t.Errorf("expected mapped_domain_id, got %s", rr.Body.String())
	}
}

func TestPersonalizeRoadmap(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/recommendation/personalize-roadmap",
		`{"domain_id":"machine_learning","completed_nodes":[]}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var out struct {
		DomainID string `json:"domain_id"`
		Path     []any  `json:"path"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if len(out.Path) != 12 {
		t.Errorf("expected 12 path nodes, got %d", len(out.Path))
	}
}

func TestPersonalizeRoadmapUnknownDomain(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/recommendation/personalize-roadmap",
		`{"domain_id":"nope","completed_nodes":[]}`)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown domain, got %d", rr.Code)
	}
}

func TestMasteryUpdate(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/mastery/update",
		`{"domain_id":"d","node_id":"n","attempt_history":[1,1,1,1,1]}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["mastered"] != true {
		t.Errorf("expected mastered with 5 correct, got %s", rr.Body.String())
	}
	if out["threshold"] != 0.95 {
		t.Errorf("expected threshold 0.95, got %v", out["threshold"])
	}
}

func TestMasteryUpdateInvalidHistory(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/mastery/update",
		`{"domain_id":"d","node_id":"n","attempt_history":[2]}`)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for invalid history, got %d", rr.Code)
	}
}

func TestMasteryIncrementalAndParams(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/mastery/update-incremental",
		`{"node_id":"n","current_p_mastery":0.5,"new_response":1}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	rr = doJSON(t, h, http.MethodGet, "/api/v1/mastery/params/mlnode", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("params status %d", rr.Code)
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["source"] != "default" {
		t.Errorf("expected default source, got %s", rr.Body.String())
	}
}

func TestAdaptiveNextAction(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/adaptive/next-action",
		`{"node_id":"x","score":0.5,"failed_attempts":2,"p_mastery":0.6}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "human_intervention") {
		t.Errorf("expected human_intervention, got %s", rr.Body.String())
	}
}

func TestGuardrailsEndpoints(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/guardrails/check",
		`{"student_text":"Just give me the answer"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["blocked"] != true || out["reason"] != "direct_answer_request" {
		t.Errorf("expected blocked direct_answer_request, got %s", rr.Body.String())
	}

	if rr := doJSON(t, h, http.MethodGet, "/api/v1/guardrails/status", ""); rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestLearningEvaluateBlocked(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/learning/evaluate",
		`{"domain_id":"machine_learning","node_id":"ml_01_programming_fundamentals","student_answer":"Write the code for me","attempt_history":null}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["blocked"] != true && out["score"] != nil {
		t.Errorf("expected guardrail block, got %s", rr.Body.String())
	}
}

func TestLearningEvaluateMissesNode(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/learning/evaluate",
		`{"domain_id":"machine_learning","node_id":"nope","student_answer":"my reasoning is fine","attempt_history":null}`)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown node, got %d", rr.Code)
	}
}

func TestLessonGeneration(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodPost, "/api/v1/learning/lesson",
		`{"domain_id":"machine_learning","node_id":"ml_01_programming_fundamentals"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var out struct {
		NodeID    string           `json:"node_id"`
		Resources []map[string]any `json:"resources"`
		Lesson    map[string]any   `json:"lesson"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out.NodeID == "" || len(out.Resources) == 0 {
		t.Errorf("expected node_id and resources, got %s", rr.Body.String())
	}
}

func TestRoadmapEndpoints(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodGet, "/api/v1/roadmap/domain-template?domain=ml_engineer", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Intro to ML") {
		t.Errorf("expected seeded ml_engineer roadmap, got %s", rr.Body.String())
	}

	rr = doJSON(t, h, http.MethodGet, "/api/v1/roadmap/list", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list status %d", rr.Code)
	}

	rr = doJSON(t, h, http.MethodPost, "/api/v1/roadmap/template",
		`{"domain":"my_domain","nodes":[{"id":"a","label":"A"}],"edges":[{"source":"a","target":"b"}]}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("post status %d body %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "my_domain") {
		t.Errorf("expected persisted record, got %s", rr.Body.String())
	}
}

func TestResourceAndLegacyEvaluate(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodGet, "/api/v1/resource?domain=software_architect&topic=system+design", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("resource status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "System Design Primer") {
		t.Errorf("expected hardcoded resource, got %s", rr.Body.String())
	}
	if rr := doJSON(t, h, http.MethodGet, "/api/v1/resource?domain=x", ""); rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 missing topic, got %d", rr.Code)
	}

	rr = doJSON(t, h, http.MethodPost, "/api/v1/evaluate", `{"user_id":"u1","domain":"software_architect","score":85}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("evaluate status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"grade":"Pass"`) {
		t.Errorf("expected Pass grade, got %s", rr.Body.String())
	}
}

func TestVoiceStatus(t *testing.T) {
	h := newTestRouter(t)
	rr := doJSON(t, h, http.MethodGet, "/api/v1/voice/status", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "asr_available") {
		t.Errorf("missing asr_available, got %s", rr.Body.String())
	}
}
