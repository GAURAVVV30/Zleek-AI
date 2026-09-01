package application

// LocalAIClientService drives the Roadmap module's AI needs with the ported
// FastAPI logic in-process (this replaces cmd/api's mockRoadmapAIClientService).

import (
	"context"
	"fmt"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/aiengine"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocalAIClientService struct {
	App     *aiengine.App
	DbPool  *pgxpool.Pool
	GroqLLM *aiengine.LLMClient
}

// GenerateRoadmapProposal maps the learner's goal text to a domain and returns
// the personalized remaining path (topo order minus completed concepts),
// mirroring FastAPI /recommendation/personalize-roadmap.
func (s *LocalAIClientService) GenerateRoadmapProposal(ctx context.Context, req AIProposalRequest) (AIProposalResponse, error) {
	var domain string
	if s.DbPool != nil && req.LearnerID != "" {
		_ = s.DbPool.QueryRow(ctx, "SELECT role FROM platform.learner_profiles WHERE user_id = $1", req.LearnerID).Scan(&domain)
	}
	if domain == "" {
		domain = s.App.Graph.MatchDomainByKeywords(req.GoalText)
		if domain == "" {
			// Last-resort: LLM mapping (same contract as goal analysis).
			llm := s.GroqLLM
			if llm == nil {
				llm = s.App.LLM
			}
			mapped := s.App.Graph.AnalyzeUserIntent(req.GoalText, llm)
			if id, ok := mapped["mapped_domain_id"].(string); ok && s.App.Graph.DomainExists(id) {
				domain = id
			}
		}
	}
	if domain == "" {
		return AIProposalResponse{}, fmt.Errorf("could not map goal text to a seeded domain")
	}

	var completed []string
	for concept, state := range req.CompetencyState {
		s := strings.ToLower(strings.TrimSpace(state))
		if s == "competent" || s == "mastered" || s == "advance" {
			completed = append(completed, concept)
		}
	}

	path, err := s.App.Graph.GetPersonalizedPath(domain, completed)
	if err != nil {
		return AIProposalResponse{}, err
	}

	items := make([]AIProposedItem, 0, len(path))
	for _, node := range path {
		nodeID, _ := node["id"].(string)
		if nodeID == "" {
			continue
		}
		items = append(items, AIProposedItem{ConceptID: nodeID})
	}
	if len(items) == 0 {
		return AIProposalResponse{}, fmt.Errorf("no remaining concepts in domain '%s' after completed nodes", domain)
	}
	return AIProposalResponse{Items: items}, nil
}

// GetConceptExplanation returns an LLM explanation of a concept, falling back
// to the embedded node data when the model is unavailable.
func (s *LocalAIClientService) GetConceptExplanation(_ context.Context, conceptID string) (*ConceptExplanation, error) {
	node := findNodeAnywhere(s.App.Graph, conceptID)
	if node == nil {
		return nil, fmt.Errorf("concept '%s' not found in any seeded domain", conceptID)
	}

	llm := s.GroqLLM
	if llm == nil {
		llm = s.App.LLM
	}

	explanation := llm.GenerateText(
		"You are a concise, friendly technical instructor.",
		fmt.Sprintf("Explain the concept '%s': %s", node.Name, conceptID),
		0.3, 250,
	)
	trimmed := strings.TrimSpace(explanation)
	if strings.HasPrefix(trimmed, "LLM unavailable") || strings.HasPrefix(trimmed, "LLM call failed") {
		trimmed = buildDeterministicExplanation(node)
	}

	prereqs, unlocks := conceptNeighbors(s.App.Graph, node)
	return &ConceptExplanation{
		ConceptID:        conceptID,
		ConceptName:      node.Name,
		Reason:           trimmed,
		PrerequisitesMet: prereqs,
		UnlocksConcepts:  unlocks,
	}, nil
}

func conceptNeighbors(engine *aiengine.GraphEngine, node *aiengine.GraphNode) (prereqs, unlocks []string) {
	for _, g := range engine.DomainGraphs {
		if _, ok := g.NodeByID[node.ID]; !ok {
			continue
		}
		for _, d := range g.HardPrereqs(node.ID) {
			if n, ok := g.NodeByID[d]; ok {
				prereqs = append(prereqs, n.Name)
			}
		}
		for _, n := range g.Nodes {
			for _, hd := range g.HardPrereqs(n.ID) {
				if hd == node.ID {
					unlocks = append(unlocks, n.Name)
					break
				}
			}
		}
		break
	}
	if prereqs == nil {
		prereqs = []string{}
	}
	if unlocks == nil {
		unlocks = []string{}
	}
	return prereqs, unlocks
}

func findNodeAnywhere(engine *aiengine.GraphEngine, nodeID string) *aiengine.GraphNode {
	for _, g := range engine.DomainGraphs {
		if n, ok := g.NodeByID[nodeID]; ok {
			return n
		}
	}
	return nil
}

func buildDeterministicExplanation(node *aiengine.GraphNode) string {
	var b strings.Builder
	b.WriteString(node.Name)
	b.WriteString(" is a core concept in this learning path. ")
	if concepts, ok := node.Raw["core_concepts"].([]any); ok && len(concepts) > 0 {
		names := make([]string, 0, len(concepts))
		for _, c := range concepts {
			if s, ok := c.(string); ok {
				names = append(names, s)
			}
		}
		if len(names) > 0 {
			b.WriteString("Core ideas include: ")
			b.WriteString(strings.Join(names, ", "))
			b.WriteString(".")
		}
	}
	if resources := node.NodeResources(); len(resources) > 0 {
		b.WriteString(" Recommended resources are listed on the concept's resource card.")
	}
	return b.String()
}
