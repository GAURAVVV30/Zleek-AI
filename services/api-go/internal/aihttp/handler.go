package aihttp

// AI service handlers — a faithful Go port of the FastAPI api/v1 routers,
// registered under the exact FastAPI /api/v1 prefixes inside the Go server.

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/aiengine"
)

const serviceVersion = "3.0"

type Handler struct {
	App *aiengine.App
}

func NewHandler() (*Handler, error) {
	app, err := aiengine.GetApp()
	if err != nil {
		return nil, err
	}
	return &Handler{App: app}, nil
}

// RegisterRoutes mirrors app.main: include_router(..., prefix="/api/v1/...").
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.handleRoot)
	mux.HandleFunc("GET /api/v1/health", h.handleHealth)

	// goal
	mux.HandleFunc("POST /api/v1/goal/analyze", h.handleGoalAnalyze)

	// roadmap (router mounted at /api/v1)
	mux.HandleFunc("GET /api/v1/roadmap", h.handleGetRoadmap)
	mux.HandleFunc("GET /api/v1/roadmap/list", h.handleListRoadmaps)
	mux.HandleFunc("POST /api/v1/roadmap", h.handlePostRoadmap)

	// resource
	mux.HandleFunc("GET /api/v1/resource", h.handleGetResource)

	// evaluate
	mux.HandleFunc("POST /api/v1/evaluate", h.handleEvaluate)

	// recommendation
	mux.HandleFunc("POST /api/v1/recommendation/personalize-roadmap", h.handlePersonalizeRoadmap)

	// learning
	mux.HandleFunc("POST /api/v1/learning/lesson", h.handleGenerateLesson)
	mux.HandleFunc("POST /api/v1/learning/evaluate", h.handleEvaluateAnswer)

	// adaptive
	mux.HandleFunc("POST /api/v1/adaptive/next-action", h.handleNextAction)

	// mastery
	mux.HandleFunc("POST /api/v1/mastery/update", h.handleMasteryUpdate)
	mux.HandleFunc("POST /api/v1/mastery/update-incremental", h.handleMasteryUpdateIncremental)
	mux.HandleFunc("GET /api/v1/mastery/params/{node_id}", h.handleMasteryParams)

	// voice
	mux.HandleFunc("GET /api/v1/voice/status", h.handleVoiceStatus)
	mux.HandleFunc("POST /api/v1/voice/transcribe", h.handleVoiceTranscribe)
	mux.HandleFunc("POST /api/v1/voice/synthesize", h.handleVoiceSynthesize)
	mux.HandleFunc("POST /api/v1/voice/tutor-session", h.handleVoiceTutorSession)

	// guardrails
	mux.HandleFunc("GET /api/v1/guardrails/status", h.handleGuardrailsStatus)
	mux.HandleFunc("POST /api/v1/guardrails/check", h.handleGuardrailsCheck)
}

// ---------------------------------------------------------------------------
// Root / health
// ---------------------------------------------------------------------------

func (h *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "AI Intelligence Service v3 — NVIDIA AI Enterprise Stack enabled",
		"stack":   "NIM + NeMo Guardrails + Riva + NV-Embed + BKT + Sentiment",
	})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	llm := h.App.LLM.GetStatus()
	voice := h.App.Voice.Status()
	guard := aiengine.GuardrailsStatus()
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "healthy",
		"service": "ai-fastapi",
		"version": serviceVersion,
		"llm":     llm,
		"voice": map[string]any{
			"asr_provider": voice["asr_provider"],
			"tts_provider": voice["tts_provider"],
		},
		"guardrails": map[string]any{
			"method":             guard["method"],
			"nvidia_key_present": guard["nvidia_key_present"],
		},
		"nvidia_enterprise_stack": llm["provider"] == "nvidia_nim",
	})
}

// ---------------------------------------------------------------------------
// Goal
// ---------------------------------------------------------------------------

func (h *Handler) handleGoalAnalyze(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserText string `json:"user_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result := h.App.Graph.AnalyzeUserIntent(req.UserText, h.App.LLM)
	// FastAPI goal.py returns analyzer output verbatim (error dict and all).
	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// Roadmap (FastAPI api/v1/roadmap.py)
// ---------------------------------------------------------------------------

func (h *Handler) handleGetRoadmap(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		domain = "software_architect"
	}
	writeJSON(w, http.StatusOK, h.App.Roadmaps.Get(domain))
}

func (h *Handler) handleListRoadmaps(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"domains": h.App.Roadmaps.List()})
}

type roadmapNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type roadmapEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func (h *Handler) handlePostRoadmap(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Domain *string       `json:"domain"`
		Nodes  []roadmapNode `json:"nodes"`
		Edges  []roadmapEdge `json:"edges"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	domain := "custom"
	if payload.Domain != nil && strings.TrimSpace(*payload.Domain) != "" {
		domain = strings.TrimSpace(*payload.Domain)
	}

	nodes := make([]map[string]any, 0, len(payload.Nodes))
	for _, n := range payload.Nodes {
		nodes = append(nodes, map[string]any{"id": n.ID, "label": n.Label})
	}
	edges := make([]map[string]any, 0, len(payload.Edges))
	for _, e := range payload.Edges {
		edges = append(edges, map[string]any{"source": e.Source, "target": e.Target})
	}

	if len(nodes) == 0 && len(edges) == 0 {
		fallback := h.App.Roadmaps.Get(domain)
		fallback["domain"] = domain
		writeJSON(w, http.StatusOK, fallback)
		return
	}

	record := aiengine.BuildRoadmapRecord(domain, nodes, edges)
	if err := h.App.Roadmaps.Put(domain, record); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to persist roadmap")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

// ---------------------------------------------------------------------------
// Resource (FastAPI api/v1/resource.py — static hardcoded payload)
// ---------------------------------------------------------------------------

func (h *Handler) handleGetResource(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	topic := r.URL.Query().Get("topic")
	if domain == "" {
		domain = "software_architect"
	}
	if topic == "" {
		writeError(w, http.StatusBadRequest, "topic is required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"domain": domain,
		"topic":  topic,
		"resource": map[string]any{
			"title":   "System Design Primer",
			"type":    "tutorial",
			"url":     "https://github.com/donnemartin/system-design-primer",
			"summary": "A curated guide for architecture, scalability, and design trade-offs.",
		},
	})
}

// ---------------------------------------------------------------------------
// Evaluate (FastAPI api/v1/evaluate.py — legacy pass/fail endpoint)
// ---------------------------------------------------------------------------

func (h *Handler) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string  `json:"user_id"`
		Domain   string  `json:"domain"`
		Score    float64 `json:"score"`
		Feedback *string `json:"feedback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Domain == "" {
		req.Domain = "software_architect"
	}
	feedback := "Keep practicing the core architecture concepts."
	if req.Feedback != nil && *req.Feedback != "" {
		feedback = *req.Feedback
	}
	grade := "Needs Improvement"
	if req.Score >= 70 {
		grade = "Pass"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":  req.UserID,
		"domain":   req.Domain,
		"score":    req.Score,
		"grade":    grade,
		"feedback": feedback,
	})
}

// ---------------------------------------------------------------------------
// Recommendation
// ---------------------------------------------------------------------------

func (h *Handler) handlePersonalizeRoadmap(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainID       string   `json:"domain_id"`
		CompletedNodes []string `json:"completed_nodes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.CompletedNodes == nil {
		req.CompletedNodes = []string{}
	}
	path, err := h.App.Graph.GetPersonalizedPath(req.DomainID, req.CompletedNodes)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"domain_id":       req.DomainID,
		"completed_nodes": req.CompletedNodes,
		"path":            path,
	})
}

// ---------------------------------------------------------------------------
// Learning
// ---------------------------------------------------------------------------

func (h *Handler) handleGenerateLesson(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainID string `json:"domain_id"`
		NodeID   string `json:"node_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	node, err := h.App.Graph.GetNode(req.DomainID, req.NodeID)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Node '%s' not found in domain '%s'", req.NodeID, req.DomainID))
		return
	}

	resources := h.App.Graph.GetResourcesForNode(req.DomainID, req.NodeID, 5)
	resourceURLs := make([]string, 0, len(resources))
	for _, res := range resources {
		if res.URL != "" {
			resourceURLs = append(resourceURLs, res.URL)
		}
	}

	conceptsText := ""
	if concepts, ok := node.Raw["key_concepts"].([]any); ok && len(concepts) > 0 {
		conceptsText = joinBullets(concepts)
	}

	systemPrompt := "You are an expert instructor. Create a concise, structured lesson that teaches the concept." +
		" Include explicit 'Key Concepts', a step-by-step explanation, and any LaTeX math expressions if relevant."
	userPrompt := fmt.Sprintf("Node ID: %s\nKey concepts:\n%s\n\nUse these supporting resources:\n%s\n\n"+
		"Return only valid JSON with: title, content_markdown, latex_expressions (array of LaTeX strings).",
		req.NodeID, conceptsText, strings.Join(resourceURLs, "\n"))

	schema := map[string]any{
		"title":             map[string]any{"type": "string"},
		"content_markdown":  map[string]any{"type": "string"},
		"latex_expressions": map[string]any{"type": "array"},
	}
	lesson := h.App.LLM.GenerateStructuredJSON(systemPrompt, userPrompt, schema)

	writeJSON(w, http.StatusOK, map[string]any{
		"node_id":   req.NodeID,
		"resources": resources,
		"lesson":    lesson,
	})
}

func (h *Handler) handleEvaluateAnswer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainID       string `json:"domain_id"`
		NodeID         string `json:"node_id"`
		StudentAnswer  string `json:"student_answer"`
		AttemptHistory []int  `json:"attempt_history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Step 1: Guardrails check (BEFORE LLM).
	guard := aiengine.CheckGuardrails(req.StudentAnswer)
	if guard.Blocked {
		writeJSON(w, http.StatusOK, map[string]any{
			"blocked":           true,
			"reason":            guard.Reason,
			"refusal_message":   guard.RefusalMessage,
			"guardrails_method": guard.Method,
			"score":             nil,
			"passed":            false,
		})
		return
	}

	// Steps 2-4: full evaluation pipeline.
	result := h.App.Graph.EvaluateSubmission(req.DomainID, req.NodeID, req.StudentAnswer, req.AttemptHistory, h.App.LLM)
	if e, ok := result["error"]; ok && e != nil {
		msg := fmt.Sprintf("%v", e)
		if strings.Contains(strings.ToLower(msg), "not found") {
			writeError(w, http.StatusNotFound, msg)
			return
		}
		// FastAPI returns the degraded error dict as a 200 alongside
		// guardrails metadata rather than raising.
		writeJSON(w, http.StatusOK, map[string]any{
			"blocked":           false,
			"guardrails_method": guard.Method,
		}, result)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"blocked":           false,
		"guardrails_method": guard.Method,
	}, result)
}

// ---------------------------------------------------------------------------
// Adaptive
// ---------------------------------------------------------------------------

func (h *Handler) handleNextAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID         string   `json:"node_id"`
		Score          float64  `json:"score"`
		FailedAttempts int      `json:"failed_attempts"`
		PMastery       *float64 `json:"p_mastery"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	out := aiengine.DetermineNextAction(req.NodeID, req.Score, req.FailedAttempts, req.PMastery)
	writeAnyJSON(w, http.StatusOK, out)
}

// ---------------------------------------------------------------------------
// Mastery (BKT)
// ---------------------------------------------------------------------------

func (h *Handler) handleMasteryUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainID       string              `json:"domain_id"`
		NodeID         string              `json:"node_id"`
		AttemptHistory []int               `json:"attempt_history"`
		CustomParams   *map[string]float64 `json:"custom_params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.AttemptHistory) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "attempt_history must contain at least one item")
		return
	}
	for _, v := range req.AttemptHistory {
		if v != 0 && v != 1 {
			writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("attempt_history must contain only 0 or 1, got %d", v))
			return
		}
	}
	var cp *aiengine.BktParams
	if req.CustomParams != nil {
		params := (*req.CustomParams)
		cp = &aiengine.BktParams{
			PInit: params["p_init"], PLearn: params["p_learn"],
			PGuess: params["p_guess"], PSlip: params["p_slip"],
		}
	}
	result := aiengine.BktEstimate(req.NodeID, req.AttemptHistory, cp)
	resp := map[string]any{"domain_id": req.DomainID, "node_id": req.NodeID}
	mergeBKT(w, resp, result)
}

func (h *Handler) handleMasteryUpdateIncremental(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID          string  `json:"node_id"`
		CurrentPMastery float64 `json:"current_p_mastery"`
		NewResponse     int     `json:"new_response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.NewResponse != 0 && req.NewResponse != 1 {
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("new_response must be 0 or 1, got %d", req.NewResponse))
		return
	}
	result := aiengine.BktEstimateIncremental(req.NodeID, req.CurrentPMastery, req.NewResponse, nil)
	mergeBKT(w, map[string]any{"node_id": req.NodeID}, result)
}

func (h *Handler) handleMasteryParams(w http.ResponseWriter, r *http.Request) {
	nodeID := r.PathValue("node_id")
	params := aiengine.GetSkillParams(nodeID)
	source := "default"
	if _, ok := aiengine.RegisteredSkills()[nodeID]; ok {
		source = "registered"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"node_id": nodeID,
		"params":  aiengine.BktParamsToList(params),
		"source":  source,
	})
}

// ---------------------------------------------------------------------------
// Voice
// ---------------------------------------------------------------------------

func (h *Handler) handleVoiceStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.App.Voice.Status())
}

func (h *Handler) handleVoiceTranscribe(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "uploaded audio file is empty")
		return
	}
	audio, err := io.ReadAll(file)
	_ = file.Close()
	if err != nil || len(audio) == 0 {
		writeError(w, http.StatusBadRequest, "Uploaded audio file is empty.")
		return
	}
	mime := r.Header.Get("Content-Type")
	if strings.HasPrefix(mime, "multipart/") {
		mime = header.Header.Get("Content-Type")
	}
	if mime == "" {
		mime = "audio/webm"
	}

	transcript, meta, err := h.App.Voice.Transcribe(audio, mime)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if strings.TrimSpace(transcript) == "" {
		writeError(w, http.StatusUnprocessableEntity, "Whisper returned an empty transcript. Check audio quality.")
		return
	}
	resp := map[string]any{"transcript": transcript, "filename": header.Filename}
	for k, v := range meta {
		resp[k] = v
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleVoiceSynthesize(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	wav, err := h.App.Voice.Synthesize(req.Text)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", "inline; filename=feedback.wav")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(wav)
}

func (h *Handler) handleVoiceTutorSession(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	domainID := r.FormValue("domain_id")
	nodeID := r.FormValue("node_id")
	attemptHistoryRaw := r.FormValue("attempt_history")
	returnAudio := true
	if ra := r.FormValue("return_audio"); ra != "" {
		if b, err := strconv.ParseBool(ra); err == nil {
			returnAudio = b
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Audio file is empty.")
		return
	}
	audio, err := io.ReadAll(file)
	_ = file.Close()
	if err != nil || len(audio) == 0 {
		writeError(w, http.StatusBadRequest, "Audio file is empty.")
		return
	}
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "audio/webm"
	}

	// Step 1: transcribe.
	transcript, asrMeta, err := h.App.Voice.Transcribe(audio, mime)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "ASR failed: "+err.Error())
		return
	}
	if strings.TrimSpace(transcript) == "" {
		writeError(w, http.StatusUnprocessableEntity, "Empty transcript — check audio quality.")
		return
	}

	// Step 2: parse optional attempt history.
	var parsedHistory []int
	if attemptHistoryRaw != "" {
		if err := json.Unmarshal([]byte(attemptHistoryRaw), &parsedHistory); err != nil {
			writeError(w, http.StatusBadRequest, "attempt_history must be valid JSON (e.g. '[1, 0, 1]').")
			return
		}
	}

	// Step 3: evaluate.
	evaluation := h.App.Graph.EvaluateSubmission(domainID, nodeID, transcript, parsedHistory, h.App.LLM)
	if e, ok := evaluation["error"]; ok && e != nil {
		msg := fmt.Sprintf("%v", e)
		if strings.Contains(strings.ToLower(msg), "not found") {
			writeError(w, http.StatusNotFound, msg)
			return
		}
		// FastAPI raises 502 for degraded (non-node) evaluation failures.
		writeError(w, http.StatusBadGateway, msg)
		return
	}

	// Step 4: synthesize spoken feedback.
	var audioBase64 string
	audioAvailable := false
	if returnAudio {
		feedback, _ := evaluation["feedback"].(string)
		pMastery := 0.0
		mastered := false
		if bkt, ok := evaluation["bkt"].(aiengine.BKTResult); ok {
			pMastery = bkt.PMastery
			mastered = bkt.Mastered
		}
		spoken := ""
		if mastered {
			spoken = fmt.Sprintf("Excellent work! You have demonstrated mastery of this concept. "+
				"Your mastery probability is now %d percent. %s", int(pMastery*100), feedback)
		} else {
			spoken = fmt.Sprintf("Your current mastery level is %d percent. %s", int(pMastery*100), feedback)
			if hint, ok := evaluation["remediation_hint"].(string); ok && hint != "" {
				spoken += " Here is a helpful hint: " + hint
			}
		}
		if wav, synthErr := h.App.Voice.Synthesize(spoken); synthErr == nil {
			audioBase64 = base64.StdEncoding.EncodeToString(wav)
			audioAvailable = true
		}
	}

	mime2 := "audio/wav"
	if !audioAvailable {
		mime2 = ""
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"transcript":      transcript,
		"asr_meta":        asrMeta,
		"evaluation":      evaluation,
		"audio_base64":    audioBase64,
		"audio_mime":      mime2,
		"audio_available": audioAvailable,
		"tts_provider":    "microsoft_speecht5_local",
	})
}

// ---------------------------------------------------------------------------
// Guardrails
// ---------------------------------------------------------------------------

func (h *Handler) handleGuardrailsStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, aiengine.GuardrailsStatus())
}

func (h *Handler) handleGuardrailsCheck(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StudentText string `json:"student_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	writeAnyJSON(w, http.StatusOK, aiengine.CheckGuardrails(req.StudentText))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, body ...map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]any{}
	for _, m := range body {
		for k, v := range m {
			resp[k] = v
		}
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]any{"detail": detail})
}

// writeAnyJSON serializes any value (struct/map) like FastAPI's JSONResponse.
func writeAnyJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func mergeBKT(w http.ResponseWriter, base map[string]any, result aiengine.BKTResult) {
	resp := base
	var params map[string]any
	if result.Params != nil {
		params = map[string]any{}
		for k, v := range result.Params {
			params[k] = v
		}
	}
	history := make([]any, 0, len(result.History))
	for _, v := range result.History {
		history = append(history, v)
	}
	resp["p_mastery"] = result.PMastery
	resp["mastered"] = result.Mastered
	resp["threshold"] = result.Threshold
	resp["attempts"] = result.Attempts
	resp["params"] = params
	resp["history"] = history
	writeJSON(w, http.StatusOK, resp)
}

func joinBullets(items []any) string {
	var b strings.Builder
	for _, item := range items {
		if s, ok := item.(string); ok {
			b.WriteString("- ")
			b.WriteString(s)
			b.WriteString("\n")
		}
	}
	return b.String()
}
