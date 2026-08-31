package aiengine

import (
	"context"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func TestMockEmbeddingGenerator(t *testing.T) {
	gen := NewMockEmbeddingGenerator(100)
	if gen.Dimension() != 100 {
		t.Errorf("Expected dimension 100, got %d", gen.Dimension())
	}

	texts := []string{"hello world", "foo bar"}
	embeddings, err := gen.GenerateEmbeddings(context.Background(), texts)
	if err != nil {
		t.Fatalf("Failed to generate embeddings: %v", err)
	}

	if len(embeddings) != 2 {
		t.Fatalf("Expected 2 embeddings, got %d", len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) != 100 {
			t.Errorf("Expected embedding %d to have length 100, got %d", i, len(emb))
		}
		// Verify normalization (L2 norm should be close to 1)
		var sumSquares float32
		for _, val := range emb {
			sumSquares += val * val
		}
		if sumSquares < 0.99 || sumSquares > 1.01 {
			t.Errorf("Embedding %d not normalized: sum of squares = %f", i, sumSquares)
		}
	}

	// Verify determinism
	emb1, _ := gen.GenerateEmbeddings(context.Background(), []string{"test string"})
	emb2, _ := gen.GenerateEmbeddings(context.Background(), []string{"test string"})
	for j := 0; j < 100; j++ {
		if emb1[0][j] != emb2[0][j] {
			t.Fatalf("Embeddings are not deterministic!")
		}
	}
}

func TestChromaRESTClientIntegration(t *testing.T) {
	client := NewChromaClient()
	ctx := context.Background()

	// 1. Get or create collection
	collectionID, err := client.GetOrCreateCollection(ctx, "test_collection_go", map[string]any{
		"description": "Integration test collection",
	})
	if err != nil {
		t.Skip("Skipping ChromaDB integration test: ChromaDB not available on port 8001 or returned error:", err)
		return
	}

	if collectionID == "" {
		t.Fatal("Collection ID is empty")
	}

	// 2. Add document
	ids := []string{"doc1", "doc2"}
	documents := []string{"Go concurrency is powerful", "Postgres index types"}
	metadatas := []map[string]any{
		{"domain_id": "test_domain", "tag": "go"},
		{"domain_id": "test_domain", "tag": "db"},
	}
	embeddings := [][]float32{
		make([]float32, 384),
		make([]float32, 384),
	}
	embeddings[0][0] = 1.0
	embeddings[1][1] = 1.0

	err = client.Add(ctx, collectionID, ids, embeddings, metadatas, documents)
	if err != nil {
		t.Fatalf("Failed to add documents to ChromaDB: %v", err)
	}

	// 3. Query document
	queryEmbedding := make([]float32, 384)
	queryEmbedding[0] = 0.9 // close to embeddings[0]
	resp, err := client.Query(ctx, collectionID, [][]float32{queryEmbedding}, 1, map[string]any{"domain_id": "test_domain"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(resp.Documents) == 0 || len(resp.Documents[0]) == 0 {
		t.Fatal("Query returned empty documents list")
	}

	bestDoc := resp.Documents[0][0]
	if !strings.Contains(bestDoc, "Go concurrency") {
		t.Errorf("Expected query to retrieve first document, got: %s", bestDoc)
	}
}

func TestSemanticRAGRetrieval(t *testing.T) {
	envMap, _ := godotenv.Read("../../../../.env")
	geminiKey := envMap["GEMINI_API_KEY"]
	if geminiKey == "" {
		t.Skip("Skipping semantic RAG retrieval test: no genuine embedding provider configured")
		return
	}
	t.Setenv("GEMINI_API_KEY", geminiKey)

	client := NewChromaClient()
	ctx := context.Background()

	collectionID, err := client.GetOrCreateCollection(ctx, "gold_standard_resources", nil)
	if err != nil {
		t.Skip("Skipping test: ChromaDB not available")
		return
	}

	// Add a specific mock document for be_01_programming_fundamentals in backend_engineer domain
	ids := []string{"test_backend_chunk"}
	documents := []string{"Clean Code principles include DRY, SOLID, and writing expressive functions."}
	metadatas := []map[string]any{
		{"domain_id": "backend_engineer", "source": "backend_engineer.pdf", "title": "Programming Fundamentals - Page 1", "page": 1},
	}
	embeddingGen := DefaultEmbeddingGenerator()
	if embeddingGen == nil {
		t.Skip("Skipping semantic RAG retrieval test: no genuine embedding provider configured")
		return
	}
	embeddings, err := embeddingGen.GenerateEmbeddings(ctx, documents)
	if err != nil {
		t.Fatalf("Failed to generate embeddings: %v", err)
	}

	err = client.Add(ctx, collectionID, ids, embeddings, metadatas, documents)
	if err != nil {
		t.Fatalf("Failed to add mock document: %v", err)
	}

	// Now run GetResourcesForNode
	// Create graph engine
	graphApp, err := GetApp()
	if err != nil {
		t.Fatalf("Failed to load GetApp: %v", err)
	}

	res := graphApp.Graph.GetResourcesForNode("backend_engineer", "be_01_programming_fundamentals", 1)
	if len(res) == 0 {
		t.Fatal("Expected RAG resources, got 0")
	}

	primary := res[0]
	if primary.Provider != "Authoritative PDF Corpus" {
		t.Errorf("Expected Provider to be 'Authoritative PDF Corpus', got %s", primary.Provider)
	}

	if !strings.Contains(primary.Title, "Programming Fundamentals - Page 1") && !strings.Contains(primary.Title, "Backend Engineer Roadmap") {
		t.Errorf("Expected Title to contain mock or real title, got: %s", primary.Title)
	}
}
