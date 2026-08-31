package infrastructure

import (
	"context"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/aiengine"
	"github.com/hcl-backend/services/api-go/internal/knowledge/application"
)

// GraphMetaProvider exposes roadmap.sh graph metadata (from the embedded
// domain graphs) for a concept node. It is the authoritative data source for
// difficulty, category, estimated effort, and prerequisite relationships.
type GraphMetaProvider struct {
	engine *aiengine.GraphEngine
	// precomputed maps
	nodeDomain map[string]string // nodeID -> domainID
	nodeName   map[string]string // nodeID -> name
}

func NewGraphMetaProvider(engine *aiengine.GraphEngine) *GraphMetaProvider {
	p := &GraphMetaProvider{
		engine:     engine,
		nodeDomain: map[string]string{},
		nodeName:   map[string]string{},
	}
	for _, domainID := range engine.DomainList {
		g := engine.DomainGraphs[domainID]
		for _, node := range g.Nodes {
			p.nodeDomain[node.ID] = domainID
			p.nodeName[node.ID] = node.Name
		}
	}
	return p
}

func (p *GraphMetaProvider) Meta(ctx context.Context, conceptID string) (*application.ConceptMeta, bool) {
	domainID, ok := p.nodeDomain[conceptID]
	if !ok {
		return nil, false
	}
	g := p.engine.DomainGraphs[domainID]
	node := g.NodeByID[conceptID]

	prereqs := []string{}
	for _, d := range g.HardPrereqs(conceptID) {
		if name := p.nodeName[d]; name != "" {
			prereqs = append(prereqs, name)
		}
	}
	successors := []string{}
	for _, n := range g.Nodes {
		for _, hd := range g.HardPrereqs(n.ID) {
			if hd == conceptID {
				if name := p.nodeName[n.ID]; name != "" {
					successors = append(successors, name)
				}
				break
			}
		}
	}

	meta := &application.ConceptMeta{
		DomainID:       domainID,
		DomainName:     g.DomainName,
		ID:             node.ID,
		Name:           node.Name,
		Category:       str(node.Raw["category"]),
		Difficulty:     str(node.Raw["difficulty_level"]),
		EstimatedHours: hours(node.Raw["estimated_hours"]),
		CoreConcepts:   strs(node.Raw["core_concepts"]),
		Prereqs:        prereqs,
		Successors:     successors,
	}
	return meta, true
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func strs(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := []string{}
	for _, a := range arr {
		if s, ok := a.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func hours(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	}
	return 0
}
