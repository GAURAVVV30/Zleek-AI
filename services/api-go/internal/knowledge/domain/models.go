package domain

import (
	"encoding/json"
	"time"
)

type Domain struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type KnowledgeStructure struct {
	ID          string    `json:"id"`
	DomainID    string    `json:"domainId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	Status      string    `json:"status"` // 'draft','published','archived'
	CreatedAt   time.Time `json:"createdAt"`
}

type Concept struct {
	ID                   string          `json:"id"`
	KnowledgeStructureID string          `json:"knowledgeStructureId"`
	Title                string          `json:"title"`
	Description          string          `json:"description"`
	LearningObjectives   json.RawMessage `json:"learningObjectives"` // JSON array of strings
	Metadata             json.RawMessage `json:"metadata"`
	CreatedAt            time.Time       `json:"createdAt"`
}

type Prerequisite struct {
	ConceptID             string `json:"conceptId"`
	PrerequisiteConceptID string `json:"prerequisiteConceptId"`
	IsHardRequirement     bool   `json:"isHardRequirement"`
}
