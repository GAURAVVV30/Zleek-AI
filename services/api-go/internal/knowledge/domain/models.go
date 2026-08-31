package domain

import (
	"time"
)

// Domain is a career role backed by a roadmap.sh graph. ID is the stable slug
// (machine_learning, software_architecture, ...).
type Domain struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	PopularGoals []string `json:"popularGoals"`
}

// KnowledgeStructure is a versioned structure of concepts for a domain.
type KnowledgeStructure struct {
	ID          string     `json:"id"`
	DomainID    string     `json:"domainId"`
	DomainName  string     `json:"domainName"`
	Version     int        `json:"version"`
	Status      string     `json:"status"` // draft|published|deprecated
	CreatedBy   string     `json:"createdBy"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// Concept is a knowledge unit. ID is the roadmap.sh node id
// (ml_01_programming_fundamentals), preserving the graph ↔ concept mapping.
type Concept struct {
	ID                   string    `json:"id"`
	KnowledgeStructureID string    `json:"knowledgeStructureId"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	CreatedAt            time.Time `json:"createdAt"`
}

// Resource is a curator-published learning resource.
type Resource struct {
	ID              string     `json:"id"`
	URL             string     `json:"url"`
	Title           string     `json:"title"`
	Author          string     `json:"author"`
	ResourceType    string     `json:"resourceType"`
	Difficulty      string     `json:"difficulty,omitempty"`
	DurationMinutes *int       `json:"durationMinutes"`
	AuthorityScore  float64    `json:"authorityScore"`
	Status          string     `json:"status"`
	CuratedBy       *string    `json:"curatedBy,omitempty"`
	CuratedAt       *time.Time `json:"curatedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// Edge is a prerequisite edge between two concepts of one structure.
type Edge struct {
	ConceptID             string `json:"conceptId"`
	PrerequisiteConceptID string `json:"prerequisiteConceptId"`
}
