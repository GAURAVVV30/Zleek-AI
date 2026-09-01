package aiengine

// LLMClient is a faithful Go port of the FastAPI app/core/llm_client.py.
//
// Provider priority (same as Python):
//  1. NVIDIA NIM  (NVIDIA_API_KEY)  -> OpenAI-compatible base URL
//  2. Groq        (GROQ_API_KEY)
//  3. Gemini      (GEMINI_API_KEY)  -> OpenAI-compatible endpoint
//
// Behavior is preserved: on NIM failure we retry Groq before giving up, and
// failures are returned as *text*, matching the Python contract.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const defaultHTTPTimeout = 60 * time.Second

// Provider detection -----------------------------------------------

type llmProvider struct {
	name  string
	key   string
	base  string
	model string
}

func detectProvider() llmProvider {
	nvidiaKey := strings.TrimSpace(os.Getenv("NVIDIA_API_KEY"))
	groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	geminiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	switch {
	case nvidiaKey != "":
		return llmProvider{"nvidia_nim", nvidiaKey, "https://integrate.api.nvidia.com/v1", "meta/llama-3.2-11b-vision-instruct"}
	case groqKey != "":
		return llmProvider{"groq", groqKey, "https://api.groq.com/openai/v1", "llama-3.3-70b-versatile"}
	case geminiKey != "":
		return llmProvider{"gemini", geminiKey, "https://generativelanguage.googleapis.com/v1beta/openai", "gemini-2.5-flash"}
	}
	return llmProvider{"none", "", "", ""}
}

// LLMClient --------------------------------------------------------

type LLMClient struct {
	provider string
	apiKey   string
	baseURL  string
	model    string
	client   *http.Client
}

func NewLLMClient(modelOverride string) *LLMClient {
	p := detectProvider()
	if e := strings.TrimSpace(os.Getenv("AI_LLM_BASE_URL")); e != "" && p.name != "none" {
		p.base = strings.TrimSuffix(e, "/")
	}
	model := p.model
	if modelOverride != "" {
		model = modelOverride
	}
	c := &LLMClient{
		provider: p.name,
		apiKey:   p.key,
		baseURL:  p.base,
		model:    model,
		client:   &http.Client{Timeout: defaultHTTPTimeout},
	}
	return c
}

func NewLLMClientForProvider(providerName, modelOverride string) *LLMClient {
	nvidiaKey := strings.TrimSpace(os.Getenv("NVIDIA_API_KEY"))
	groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	geminiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))

	var p llmProvider
	switch strings.ToLower(providerName) {
	case "nvidia", "nvidia_nim":
		if nvidiaKey != "" {
			p = llmProvider{"nvidia_nim", nvidiaKey, "https://integrate.api.nvidia.com/v1", "meta/llama-3.2-11b-vision-instruct"}
		}
	case "groq":
		if groqKey != "" {
			p = llmProvider{"groq", groqKey, "https://api.groq.com/openai/v1", "llama-3.3-70b-versatile"}
		}
	case "gemini":
		if geminiKey != "" {
			p = llmProvider{"gemini", geminiKey, "https://generativelanguage.googleapis.com/v1beta/openai", "gemini-2.5-flash"}
		}
	}

	if p.name == "" {
		p = detectProvider()
	}

	if e := strings.TrimSpace(os.Getenv("AI_LLM_BASE_URL")); e != "" && p.name != "none" {
		p.base = strings.TrimSuffix(e, "/")
	}
	model := p.model
	if modelOverride != "" {
		model = modelOverride
	}
	return &LLMClient{
		provider: p.name,
		apiKey:   p.key,
		baseURL:  p.base,
		model:    model,
		client:   &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// DefaultLLM returns the shared singleton mirroring Python's module-level use.
func DefaultLLM() *LLMClient {
	return NewLLMClient("")
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (l *LLMClient) callChat(baseURL, apiKey, model, systemPrompt, userPrompt string, temperature float64, maxTokens int) (string, error) {
	payload := chatRequest{
		Model:       model,
		Messages:    []chatMessage{{Role: "system", Content: systemPrompt}, {Role: "user", Content: userPrompt}},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, strings.TrimSuffix(baseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := l.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("chat completions HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("no choices in chat completions response")
	}
	return parsed.Choices[0].Message.Content, nil
}

// GenerateText mirrors generate_text(): returns text or a failure description.
func (l *LLMClient) GenerateText(systemPrompt, userPrompt string, temperature float64, maxTokens int) string {
	if l.provider == "none" {
		return "LLM unavailable: no API key configured (NVIDIA_API_KEY or GROQ_API_KEY)."
	}
	if l.provider == "nvidia_nim" || l.provider == "groq" || l.provider == "gemini" {
		text, err := l.callChat(l.baseURL, l.apiKey, l.model, systemPrompt, userPrompt, temperature, maxTokens)
		if err == nil {
			return text
		}
		// Emergency fallback to Groq, exactly like Python.
		if l.provider == "nvidia_nim" {
			if groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY")); groqKey != "" {
				if text, fallbackErr := l.callChat("https://api.groq.com/openai/v1", groqKey, "llama-3.3-70b-versatile", systemPrompt, userPrompt, temperature, maxTokens); fallbackErr == nil {
					return text
				}
			}
		}
		return fmt.Sprintf("LLM call failed: %v", err)
	}
	return "LLM unavailable: unknown provider"
}

// GenerateStructuredJSON mirrors generate_structured_json().
func (l *LLMClient) GenerateStructuredJSON(systemPrompt, userPrompt string, responseSchema map[string]any) map[string]any {
	schemaInstruction := ""
	if len(responseSchema) > 0 {
		enc, _ := json.Marshal(responseSchema)
		schemaInstruction = "Respond ONLY with valid JSON matching this schema: " + string(enc) + "\n"
	}
	fullUserPrompt := schemaInstruction + userPrompt
	raw := l.GenerateText(systemPrompt, fullUserPrompt, 0.3, 1024)

	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out
	}
	// Extract the first JSON object (handles markdown code fences).
	m := jsonObjectRE.FindString(raw)
	if m != "" {
		if err := json.Unmarshal([]byte(m), &out); err == nil {
			return out
		} else {
			return map[string]any{"error": fmt.Sprintf("Failed to parse JSON: %v", err), "raw": raw}
		}
	}
	return map[string]any{"error": "Model did not return valid JSON.", "raw": raw}
}

var jsonObjectRE = regexp.MustCompile(`(?s)\{.*\}`)

// GetStatus mirrors get_status() for health checks.
func (l *LLMClient) GetStatus() map[string]any {
	return map[string]any{
		"provider":        l.provider,
		"model":           l.model,
		"api_key_present": l.apiKey != "",
	}
}

func (l *LLMClient) Provider() string { return l.provider }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
