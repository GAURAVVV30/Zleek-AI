package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ChromaClient struct {
	baseURL string
	client  *http.Client
}

func NewChromaClient() *ChromaClient {
	url := strings.TrimSpace(os.Getenv("CHROMA_URL"))
	if url == "" {
		url = "http://localhost:8001"
	}
	url = strings.TrimSuffix(url, "/")
	return &ChromaClient{
		baseURL: url + "/api/v2",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type collectionRequest struct {
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type collectionResponse struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata"`
}

func (c *ChromaClient) GetOrCreateCollection(ctx context.Context, name string, metadata map[string]any) (string, error) {
	reqBody := collectionRequest{
		Name:     name,
		Metadata: metadata,
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.baseURL + "/tenants/default_tenant/databases/default_database/collections"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		// If it already exists, let's get it by name.
		if strings.Contains(string(respBody), "already exists") {
			return c.getCollectionID(ctx, name)
		}
		return "", fmt.Errorf("chroma collection creation returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var res collectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.ID, nil
}

func (c *ChromaClient) getCollectionID(ctx context.Context, name string) (string, error) {
	url := fmt.Sprintf("%s/tenants/default_tenant/databases/default_database/collections/%s", c.baseURL, name)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("chroma get collection returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var res collectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.ID, nil
}

type addRequest struct {
	IDs        []string         `json:"ids"`
	Embeddings [][]float32      `json:"embeddings"`
	Metadatas  []map[string]any `json:"metadatas"`
	Documents  []string         `json:"documents"`
}

func (c *ChromaClient) Add(ctx context.Context, collectionID string, ids []string, embeddings [][]float32, metadatas []map[string]any, documents []string) error {
	reqBody := addRequest{
		IDs:        ids,
		Embeddings: embeddings,
		Metadatas:  metadatas,
		Documents:  documents,
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/tenants/default_tenant/databases/default_database/collections/%s/add", c.baseURL, collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma add returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type queryRequest struct {
	QueryEmbeddings [][]float32    `json:"query_embeddings"`
	NResults        int            `json:"n_results"`
	Where           map[string]any `json:"where,omitempty"`
	Include         []string       `json:"include"`
}

type QueryResponse struct {
	IDs       [][]string         `json:"ids"`
	Documents [][]string         `json:"documents"`
	Metadatas [][]map[string]any `json:"metadatas"`
	Distances [][]float32        `json:"distances"`
}

func (c *ChromaClient) Query(ctx context.Context, collectionID string, queryEmbeddings [][]float32, nResults int, where map[string]any) (*QueryResponse, error) {
	reqBody := queryRequest{
		QueryEmbeddings: queryEmbeddings,
		NResults:        nResults,
		Where:           where,
		Include:         []string{"documents", "metadatas", "distances"},
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/tenants/default_tenant/databases/default_database/collections/%s/query", c.baseURL, collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chroma query returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var res QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *ChromaClient) DeleteCollection(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/tenants/default_tenant/databases/default_database/collections/%s", c.baseURL, name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma delete collection returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
