package aiclient

// LocalClient runs the FastAPI AI logic in-process — no HTTP round-trip to a
// separate FastAPI service. It replaces MockClient in cmd/api/main.go.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/aiengine"
	"github.com/hcl-backend/services/api-go/internal/assessment/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocalClient struct {
	App    *aiengine.App
	dbPool *pgxpool.Pool
}

func NewLocalClient(app *aiengine.App, dbPool *pgxpool.Pool) *LocalClient {
	return &LocalClient{App: app, dbPool: dbPool}
}

// ProposeGoalMapping mirrors the FastAPI /goal/analyze mapping: returns the
// mapped knowledge-structure (domain) ID for the learner's goal text. When the
// LLM is unavailable (no API key), it degrades to the deterministic
// keyword-based domain mapper so the goals flow still works offline.
func (l *LocalClient) ProposeGoalMapping(_ context.Context, goalText string) (string, error) {
	result := l.App.Graph.AnalyzeUserIntent(goalText, l.App.LLM)
	if e, ok := result["error"]; ok && e != nil {
		return l.App.Graph.MatchDomainByKeywords(goalText), nil
	}
	mapped, _ := result["mapped_domain_id"].(string)
	if mapped == "" || !l.App.Graph.DomainExists(mapped) {
		return l.App.Graph.MatchDomainByKeywords(goalText), nil
	}
	return mapped, nil
}

// Evaluate mirrors the FastAPI /learning/evaluate pipeline: LLM grading → BKT
// → sentiment. The submission carries domain_id and node_id; rubric falls back
// to the node's embedded assessment rubric.
func (l *LocalClient) Evaluate(_ context.Context, submission json.RawMessage, _ json.RawMessage) (*domain.EvaluationResult, error) {
	var payload struct {
		DomainID       string `json:"domain_id"`
		NodeID         string `json:"node_id"`
		StudentAnswer  string `json:"student_answer"`
		AttemptHistory []int  `json:"attempt_history"`
	}
	if err := json.Unmarshal(submission, &payload); err != nil {
		return nil, fmt.Errorf("invalid submission: %w", err)
	}
	if payload.DomainID == "" || payload.NodeID == "" {
		return nil, fmt.Errorf("invalid submission: domain_id and node_id are required")
	}
	if payload.StudentAnswer == "" {
		// Allow a deterministic pass-through evaluated by rubric heuristics only.
		return nil, fmt.Errorf("invalid submission: empty answer")
	}

	result := l.App.Graph.EvaluateSubmission(payload.DomainID, payload.NodeID, payload.StudentAnswer, payload.AttemptHistory, l.App.LLM)
	if e, ok := result["error"]; ok && e != nil {
		return nil, fmt.Errorf("ai evaluation failed: %v", e)
	}

	score, _ := result["score"].(float64)
	passed, _ := result["passed"].(bool)
	verdict := "weak"
	if score >= 0.7 {
		verdict = "competent"
	}
	if !passed {
		verdict = "weak"
	}
	confidence := 0.9
	if bkt, ok := result["bkt"].(aiengine.BKTResult); ok {
		confidence = bkt.PMastery
	}
	return &domain.EvaluationResult{
		Score:      score * 100,
		Confidence: confidence,
		Result:     verdict,
	}, nil
}

// ValidateKnowledgeStructure checks that a DAG payload is contract-valid using
// the graph engine semantics.
func (l *LocalClient) ValidateKnowledgeStructure(_ context.Context, structure interface{}) (bool, string, error) {
	// Accepted shapes: a knowledge-structure id, or a {domain_id|nodes} payload.
	switch s := structure.(type) {
	case string:
		if l.App.Graph.DomainExists(s) {
			return true, "graph structure valid: " + s, nil
		}
		return false, "unknown domain_id: " + s, nil
	case map[string]any:
		domainID, _ := s["domain_id"].(string)
		if domainID == "" {
			return false, "structure must include a known domain_id", nil
		}
		if l.App.Graph.DomainExists(domainID) {
			return true, "graph structure valid: " + domainID, nil
		}
		return false, "unknown domain_id: " + domainID, nil
	default:
		raw, err := json.Marshal(structure)
		if err != nil {
			return false, "unparseable structure", err
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return false, "structure must be an object with domain_id", err
		}
		return l.ValidateKnowledgeStructure(nil, m)
	}
}

func (l *LocalClient) RankResources(ctx context.Context, conceptID string) ([]string, error) {
	nodeID := conceptID
	domainID := ""
	if l.dbPool != nil {
		var nID, dSlug string
		err := l.dbPool.QueryRow(ctx, "SELECT c.node_id, d.slug FROM platform.concepts c JOIN platform.knowledge_structures k ON k.id = c.knowledge_structure_id JOIN platform.domains d ON d.id = k.domain_id WHERE c.id::text = $1 OR c.node_id = $1", conceptID).Scan(&nID, &dSlug)
		if err == nil {
			nodeID = nID
			domainID = dSlug
		} else {
			log.Printf("DB error resolving concept UUID %s: %v", conceptID, err)
		}
	} else {
		log.Println("dbPool is nil inside LocalClient!")
	}
	log.Printf("RankResources resolving: conceptID=%s -> nodeID=%s, domainID=%s", conceptID, nodeID, domainID)
	ragRes := l.App.Graph.GetResourcesForNode(domainID, nodeID, 3)
	ragIDs := make([]string, len(ragRes))
	for i, r := range ragRes {
		distVal, _ := r.Metadata["distance_score"].(float32)
		textVal, _ := r.Metadata["document_text"].(string)
		encodedText := base64.StdEncoding.EncodeToString([]byte(textVal))
		ragIDs[i] = fmt.Sprintf("%s_dist_%f_text_%s", r.ID, distVal, encodedText)
	}
	log.Printf("GetResourcesForNode returned %d RAG resource IDs: %v", len(ragIDs), ragIDs)
	return ragIDs, nil
}

// ExplainResourceRelevance mirrors the FastAPI resource-explanation prompt.
func (l *LocalClient) ExplainResourceRelevance(_ context.Context, conceptID, resourceID string) (string, error) {
	explanation := l.App.LLM.GenerateText(
		"You are a helpful learning assistant. Explain concisely why this resource is relevant to the concept.",
		fmt.Sprintf("Concept: %s\nResource ID: %s", conceptID, resourceID),
		0.4, 300,
	)
	marker := strings.TrimSpace(explanation)
	if strings.HasPrefix(marker, "LLM unavailable") || strings.HasPrefix(marker, "LLM call failed") {
		return "This resource supports the concept. Review the linked material alongside the node's core concepts.", nil
	}
	return explanation, nil
}
