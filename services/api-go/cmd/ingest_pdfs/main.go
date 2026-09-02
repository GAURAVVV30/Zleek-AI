package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/hcl-backend/services/api-go/internal/aiengine"
)

var pdfToDomain = map[string]string{
	"ai-data-scientist.pdf":  "ai_data_scientist",
	"ai-engineer.pdf":        "ai_engineer",
	"backend_engineer.pdf":   "backend_engineer",
	"data-analyst.pdf":       "data_analyst",
	"devops_sre.pdf":         "devops_sre",
	"frontend_enginner.pdf":  "frontend_engineer",
	"full-stack.pdf":         "full_stack",
	"machine-learning.pdf":   "machine_learning",
	"Mobile_Engineer.pdf":    "mobile_engineer",
	"product-manager.pdf":    "product_manager",
	"software-architect.pdf": "software_architect",
}

func main() {
	_ = godotenv.Load("../../.env")
	log.Println("Starting Go PDF RAG Ingestion Pipeline...")

	pdfDir := "/home/zuzu/test/raw_pdfs"
	if _, err := os.Stat(pdfDir); os.IsNotExist(err) {
		log.Fatalf("PDF directory not found at %s", pdfDir)
	}

	chromaClient := aiengine.NewChromaClient()
	embeddingGen := aiengine.DefaultEmbeddingGenerator()
	if embeddingGen == nil {
		log.Fatal("No genuine embedding provider configured. Ingestion aborted.")
	}

	ctx := context.Background()
	log.Println("Purging stale gold_standard_resources collection...")
	_ = chromaClient.DeleteCollection(ctx, "gold_standard_resources")

	collectionID, err := chromaClient.GetOrCreateCollection(ctx, "gold_standard_resources", map[string]any{
		"description": "Seeded authoritative learning PDF corpus chunks",
	})
	if err != nil {
		log.Fatalf("Failed to get/create ChromaDB collection: %v", err)
	}
	log.Printf("Using ChromaDB collection ID: %s", collectionID)

	totalChunks := 0

	for filename, domainID := range pdfToDomain {
		pdfPath := filepath.Join(pdfDir, filename)
		if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
			log.Printf("Warning: PDF file %s not found. Skipping.", pdfPath)
			continue
		}

		log.Printf("Processing %s for domain %s...", filename, domainID)

		roleName := strings.Title(strings.ReplaceAll(domainID, "_", " "))
		page := 1
		for {
			cmd := exec.Command("pdftotext", "-f", fmt.Sprint(page), "-l", fmt.Sprint(page), pdfPath, "-")
			var out bytes.Buffer
			cmd.Stdout = &out
			err := cmd.Run()
			if err != nil || len(strings.TrimSpace(out.String())) == 0 {
				break // End of PDF or error
			}

			pageText := out.String()
			chunks := chunkText(pageText, 800, 150)

			var ids []string
			var documents []string
			var metadatas []map[string]any

			for idx, chunk := range chunks {
				chunkID := fmt.Sprintf("%s_page%d_chunk%d", domainID, page, idx)
				title := fmt.Sprintf("%s Roadmap - Page %d", roleName, page)

				ids = append(ids, chunkID)
				documents = append(documents, chunk)
				metadatas = append(metadatas, map[string]any{
					"domain_id":   domainID,
					"source":      filename,
					"title":       title,
					"page":        page,
					"chunk_index": idx,
					"role":        roleName,
					"format":      "article",
				})
			}

			if len(ids) > 0 {
				embeddings, err := embeddingGen.GenerateEmbeddings(ctx, documents)
				if err != nil {
					log.Fatalf("Failed to generate embeddings: %v", err)
				}

				err = chromaClient.Add(ctx, collectionID, ids, embeddings, metadatas, documents)
				if err != nil {
					log.Fatalf("Failed to index chunks into ChromaDB: %v", err)
				}
				totalChunks += len(ids)
			}

			page++
		}
		log.Printf("Successfully ingested domain %s: %d pages", domainID, page-1)
	}

	log.Printf("PDF RAG Ingestion Complete! Total indexed chunks: %d", totalChunks)
}

func chunkText(text string, chunkSize, overlap int) []string {
	var chunks []string
	runes := []rune(text)
	if len(runes) == 0 {
		return chunks
	}
	if len(runes) <= chunkSize {
		return []string{strings.TrimSpace(text)}
	}
	for i := 0; i < len(runes); {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, strings.TrimSpace(string(runes[i:end])))
		if end == len(runes) {
			break
		}
		i += chunkSize - overlap
	}
	return chunks
}
