package aiengine

import "embed"

// domainFS embeds the FastAPI knowledge base: every domain's *_graph.json,
// skill graphs, assessments, and RAG resource indexes, plus the roadmap store.
//
//go:embed all:data
var dataFS embed.FS

const (
	domainDataRoot  = "data/domains"
	roadmapDataFile = "data/roadmaps.json"
)
