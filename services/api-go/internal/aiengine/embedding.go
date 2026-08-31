package aiengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type EmbeddingGenerator interface {
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	Dimension() int
}

type NvidiaEmbeddingGenerator struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewNvidiaEmbeddingGenerator(apiKey string) *NvidiaEmbeddingGenerator {
	baseURL := "https://integrate.api.nvidia.com/v1"
	if e := strings.TrimSpace(os.Getenv("AI_LLM_BASE_URL")); e != "" {
		baseURL = strings.TrimSuffix(e, "/")
	}
	return &NvidiaEmbeddingGenerator{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   "nvidia/nv-embed-v1",
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type nvidiaEmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	InputType      string   `json:"input_type,omitempty"`
	EncodingFormat string   `json:"encoding_format"`
	Truncate       string   `json:"truncate,omitempty"`
}

type nvidiaEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

func (g *NvidiaEmbeddingGenerator) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := nvidiaEmbeddingRequest{
		Model:          g.model,
		Input:          texts,
		InputType:      "passage",
		EncodingFormat: "float",
		Truncate:       "END",
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/embeddings", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nvidia embeddings NIM API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var res nvidiaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(texts))
	for _, d := range res.Data {
		if d.Index < len(embeddings) {
			embeddings[d.Index] = d.Embedding
		}
	}
	return embeddings, nil
}

func (g *NvidiaEmbeddingGenerator) Dimension() int {
	return 4096
}

type GeminiEmbeddingGenerator struct {
	apiKey string
	client *http.Client
}

func NewGeminiEmbeddingGenerator(apiKey string) *GeminiEmbeddingGenerator {
	return &GeminiEmbeddingGenerator{
		apiKey: apiKey,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiEmbedRequest struct {
	Model   string        `json:"model"`
	Content geminiContent `json:"content"`
}

type geminiBatchRequest struct {
	Requests []geminiEmbedRequest `json:"requests"`
}

type geminiBatchResponse struct {
	Embeddings []struct {
		Values []float32 `json:"values"`
	} `json:"embeddings"`
}

func (g *GeminiEmbeddingGenerator) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := geminiBatchRequest{Requests: make([]geminiEmbedRequest, len(texts))}
	for i, t := range texts {
		reqBody.Requests[i] = geminiEmbedRequest{
			Model: "models/gemini-embedding-001",
			Content: geminiContent{
				Parts: []geminiPart{{Text: t}},
			},
		}
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:batchEmbedContents?key=" + g.apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini batchEmbedContents API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var res geminiBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(texts))
	for i, e := range res.Embeddings {
		if i < len(embeddings) {
			embeddings[i] = e.Values
		}
	}
	return embeddings, nil
}

func (g *GeminiEmbeddingGenerator) Dimension() int {
	return 3072
}

type MockEmbeddingGenerator struct {
	dimension int
}

func NewMockEmbeddingGenerator(dim int) *MockEmbeddingGenerator {
	return &MockEmbeddingGenerator{dimension: dim}
}

func (g *MockEmbeddingGenerator) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for idx, text := range texts {
		emb := make([]float32, g.dimension)
		var sum float64
		for i := 0; i < g.dimension; i++ {
			h := sha256.New()
			h.Write([]byte(fmt.Sprintf("%s:%d", text, i)))
			hashBytes := h.Sum(nil)
			val := float32(binary.BigEndian.Uint32(hashBytes[:4])) / float32(1<<32-1)
			emb[i] = val
			sum += float64(val * val)
		}
		norm := float32(math.Sqrt(sum))
		if norm > 0 {
			for i := 0; i < g.dimension; i++ {
				emb[i] /= norm
			}
		}
		embeddings[idx] = emb
	}
	return embeddings, nil
}

func (g *MockEmbeddingGenerator) Dimension() int {
	return g.dimension
}

func DefaultEmbeddingGenerator() EmbeddingGenerator {
	if key := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); key != "" {
		return NewGeminiEmbeddingGenerator(key)
	}
	if key := strings.TrimSpace(os.Getenv("NVIDIA_API_KEY")); key != "" {
		return NewNvidiaEmbeddingGenerator(key)
	}
	// Explicitly return nil when no genuine provider is configured, to satisfy gap requirement
	return nil
}
