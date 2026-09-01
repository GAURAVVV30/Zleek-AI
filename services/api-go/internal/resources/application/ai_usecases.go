package application

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/resources/domain"
)

type AIClient interface {
	RankResources(ctx context.Context, conceptID string) ([]string, error)
	ExplainResourceRelevance(ctx context.Context, conceptID, resourceID string) (string, error)
}

type GetAlternateResourcesUseCase struct {
	repo     ResourceRepository
	aiClient AIClient
}

func NewGetAlternateResourcesUseCase(repo ResourceRepository, aiClient AIClient) *GetAlternateResourcesUseCase {
	return &GetAlternateResourcesUseCase{repo: repo, aiClient: aiClient}
}

func (uc *GetAlternateResourcesUseCase) Execute(ctx context.Context, conceptID string) ([]domain.Resource, error) {
	// 1. Ask AI to rank/suggest alternate resources for this concept
	resourceIDs, err := uc.aiClient.RankResources(ctx, conceptID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch the actual resources from DB
	var resources []domain.Resource
	for _, rawID := range resourceIDs {
		id := rawID
		var distStr string
		var docText string
		if strings.Contains(rawID, "_dist_") {
			parts := strings.Split(rawID, "_dist_")
			id = parts[0]
			if len(parts) > 1 {
				subParts := strings.Split(parts[1], "_text_")
				distStr = subParts[0]
				if len(subParts) > 1 {
					if decoded, err := base64.StdEncoding.DecodeString(subParts[1]); err == nil {
						docText = string(decoded)
					}
				}
			}
		}

		res, err := uc.repo.GetResource(ctx, id)
		if err == nil {
			resources = append(resources, *res)
		} else {
			// If not found in DB, check if it's a PDF chunk from Chroma RAG
			if strings.Contains(id, "_page") && strings.Contains(id, "_chunk") {
				parts := strings.Split(id, "_page")
				domainSlug := parts[0]
				
				pageNum := "1"
				pageParts := strings.Split(parts[1], "_chunk")
				if len(pageParts) > 0 {
					pageNum = pageParts[0]
				}
				
				sourceFile := domainSlug + ".pdf"
				if domainSlug == "frontend_engineer" {
					sourceFile = "frontend_enginner.pdf"
				} else if domainSlug == "mobile_engineer" {
					sourceFile = "Mobile_Engineer.pdf"
				} else if domainSlug == "full_stack" {
					sourceFile = "full-stack.pdf"
				} else if domainSlug == "data_engineer" {
					sourceFile = "data-engineer.pdf"
				} else if domainSlug == "product_manager" {
					sourceFile = "product-manager.pdf"
				} else if domainSlug == "data_analyst" {
					sourceFile = "data-analyst.pdf"
				} else if domainSlug == "ai_data_scientist" {
					sourceFile = "ai-data-scientist.pdf"
				} else if domainSlug == "ai_engineer" {
					sourceFile = "ai-engineer.pdf"
				} else if domainSlug == "software_architect" || domainSlug == "software_architecture" {
					sourceFile = "software-architect.pdf"
				} else if domainSlug == "machine_learning" {
					sourceFile = "machine-learning.pdf"
				}
				
				author := "Authoritative PDF Corpus"
				resType := "article"
				score := 1.0
				
				var provenanceNote *string
				if distStr != "" {
					pNote := fmt.Sprintf("distance_score:%s|chunk_text:%s", distStr, docText)
					provenanceNote = &pNote
				}
				
				resources = append(resources, domain.Resource{
					ID:             id,
					URL:            fmt.Sprintf("file:///home/zuzu/test/raw_pdfs/%s#page=%s", sourceFile, pageNum),
					Author:         &author,
					ResourceType:   resType,
					AuthorityScore: &score,
					Status:         domain.StatusPublished,
					ProvenanceNote: provenanceNote,
				})
			}
		}
	}
	if resources == nil {
		resources = []domain.Resource{}
	}
	return resources, nil
}

type ExplainResourceRelevanceUseCase struct {
	aiClient AIClient
}

func NewExplainResourceRelevanceUseCase(aiClient AIClient) *ExplainResourceRelevanceUseCase {
	return &ExplainResourceRelevanceUseCase{aiClient: aiClient}
}

func (uc *ExplainResourceRelevanceUseCase) Execute(ctx context.Context, conceptID, resourceID string) (string, error) {
	return uc.aiClient.ExplainResourceRelevance(ctx, conceptID, resourceID)
}
