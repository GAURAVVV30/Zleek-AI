package aiengine

// RagEngine is a Go port of app/core/rag_engine.py.
//
// ChromaDB + NVIDIA NV-Embed are integrated in Go.
// If ChromaDB vector-retrieval fails or is unavailable, the deterministic fallback
// (using local domain graphs and resources_rag.json) is used to maintain backward compatibility.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// RAGResource is a {id, title, url, provider, metadata} result, matching the
// shape ResourceRetriever.get_resources_for_node() returns.
type RAGResource struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	URL            string         `json:"url"`
	Provider       string         `json:"provider"`
	Metadata       map[string]any `json:"metadata"`
	EmbeddingModel string         `json:"embedding_model"`
}

// embeddingModel mirrors NV_EMBED_MODEL when NV-Embed is active.
const ragEmbeddingModel = "local_tag_fallback"

// GetResourcesForNode mirrors ResourceRetriever.get_resources_for_node() using
// semantic vector search in ChromaDB. Falls back to deterministic tag-filtering if Chroma fails.
func (engine *GraphEngine) GetResourcesForNode(domainID, nodeID string, nResults int) []RAGResource {
	results := []RAGResource{}
	node, err := engine.GetNode(domainID, nodeID)
	queryText := nodeID
	if err == nil && node != nil {
		queryText = node.Name
	}

	// Try semantic ChromaDB search
	embeddingGen := DefaultEmbeddingGenerator()
	if embeddingGen != nil {
		chromaClient := NewChromaClient()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if collectionID, err := chromaClient.getCollectionID(ctx, "gold_standard_resources"); err == nil && collectionID != "" {
			queryEmbeddings, err := embeddingGen.GenerateEmbeddings(ctx, []string{queryText})
			if err == nil && len(queryEmbeddings) > 0 {
				var where map[string]any
				if domainID != "" {
					chromaDomainID := domainID
					if domainID == "software_architecture" {
						chromaDomainID = "software_architect"
					}
					where = map[string]any{"domain_id": chromaDomainID}
				}
				resp, err := chromaClient.Query(ctx, collectionID, queryEmbeddings, nResults, where)
				if err == nil && resp != nil && len(resp.Documents) > 0 && len(resp.Documents[0]) > 0 {
					modelName := "chromadb_default"
					if _, ok := embeddingGen.(*NvidiaEmbeddingGenerator); ok {
						modelName = "nvidia/nv-embed-v1"
					} else if _, ok := embeddingGen.(*GeminiEmbeddingGenerator); ok {
						modelName = "models/text-embedding-004"
					}

					docs := resp.Documents[0]
					metas := resp.Metadatas[0]
					ids := resp.IDs[0]

					for i := 0; i < len(docs); i++ {
						resID := ids[i]
						doc := docs[i]
						meta := metas[i]

						title, _ := meta["title"].(string)
						if title == "" {
							title = doc
							if len(title) > 60 {
								title = title[:57] + "..."
							}
						}
						source, _ := meta["source"].(string)
						page, _ := meta["page"].(float64)

						url := fmt.Sprintf("file:///home/zuzu/test/raw_pdfs/%s#page=%d", source, int(page))
						meta["node_id"] = nodeID
						meta["type"] = "article"
						meta["authority_score"] = 1.0
						if len(resp.Distances) > 0 && i < len(resp.Distances[0]) {
							meta["distance_score"] = resp.Distances[0][i]
						}
						if len(resp.Documents) > 0 && i < len(resp.Documents[0]) {
							meta["document_text"] = resp.Documents[0][i]
						}

						results = append(results, RAGResource{
							ID:             resID,
							Title:          "[PDF] " + title,
							URL:            url,
							Provider:       "Authoritative PDF Corpus",
							Metadata:       meta,
							EmbeddingModel: modelName,
						})
					}
					return results
				}
			}
		}
	} else {
		log.Printf("ChromaDB semantic query unavailable: no genuine embedding provider configured. Falling back to local tag-matching.")
	}

	// Fallback logic in case ChromaDB is down or not seeded yet
	log.Printf("ChromaDB semantic query unavailable for node %s. Falling back to local tag-matching.", nodeID)

	if node == nil {
		// Try a cross-domain lookup by node ID
		if cross := findNodeAcrossDomains(engine, nodeID); cross != nil {
			node = cross
		} else {
			return results
		}
	}

	for idx, res := range node.NodeResources() {
		if len(results) >= nResults {
			break
		}
		title, _ := res["title"].(string)
		url, _ := res["url"].(string)
		provider, _ := res["provider"].(string)
		id := nodeResID(nodeID, idx)
		if rawID, ok := res["id"].(string); ok && rawID != "" {
			id = rawID
		}
		meta := map[string]any{"node_id": nodeID}
		for _, k := range []string{"type", "provider", "authority_score"} {
			if v, ok := res[k]; ok {
				meta[k] = v
			}
		}
		results = append(results, RAGResource{
			ID:             id,
			Title:          title,
			URL:            url,
			Provider:       provider,
			Metadata:       meta,
			EmbeddingModel: ragEmbeddingModel,
		})
	}

	ragDomains := []string{}
	if engine.DomainExists(domainID) {
		ragDomains = append(ragDomains, domainID)
	} else {
		ragDomains = engine.DomainList
	}
	for _, d := range ragDomains {
		if len(results) >= nResults {
			break
		}
		rag := LoadRagResources(d)
		resources, _ := rag["resources"].([]any)
		for _, r := range resources {
			if len(results) >= nResults {
				break
			}
			rm, ok := r.(map[string]any)
			if !ok {
				continue
			}
			id, _ := rm["id"].(string)
			if id == "" {
				continue
			}
			tags, _ := rm["tags"].([]any)
			if !tagMatches(tags, nodeID) {
				continue
			}
			title, _ := rm["title"].(string)
			url, _ := rm["url"].(string)
			results = append(results, RAGResource{
				ID:             id,
				Title:          title,
				URL:            url,
				Metadata:       map[string]any{"node_id": nodeID, "tags": tags},
				EmbeddingModel: ragEmbeddingModel,
			})
		}
	}

	if results == nil {
		results = []RAGResource{}
	}
	return results
}

func tagMatches(tags []any, nodeID string) bool {
	if len(tags) == 0 {
		return true
	}
	for _, t := range tags {
		s, ok := t.(string)
		if !ok {
			continue
		}
		ns := strings.ReplaceAll(nodeID, "_", " ")
		if strings.Contains(strings.ToLower(s), strings.ToLower(ns)) ||
			strings.Contains(strings.ToLower(nodeID), strings.ToLower(s)) {
			return true
		}
	}
	return false
}

func findNodeAcrossDomains(engine *GraphEngine, nodeID string) *GraphNode {
	for _, g := range engine.DomainGraphs {
		if n, ok := g.NodeByID[nodeID]; ok {
			return n
		}
	}
	return nil
}

func nodeResID(nodeID string, idx int) string {
	return fmt.Sprintf("node:%s:%d", nodeID, idx)
}

func LoadRagResources(domain string) map[string]any {
	path := domainDataRoot + "/" + domain + "/resources_rag.json"
	raw, err := dataFS.ReadFile(path)
	if err != nil {
		return map[string]any{"domain": domain, "resources": []any{}}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{"domain": domain, "resources": []any{}}
	}
	if payload["resources"] == nil {
		payload["resources"] = []any{}
	}
	return payload
}

func (engine *GraphEngine) RankNodeResources(nodeID string, n int) ([]string, error) {
	res := engine.GetResourcesForNode("", nodeID, n)
	ids := make([]string, 0, len(res))
	for _, r := range res {
		ids = append(ids, r.ID)
	}
	return ids, nil
}
