package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/diagnostics/domain"
)

// SessionStore persists diagnostic sessions (in-memory is sufficient for the
// short-lived onboarding flow; the store is swappable).
type SessionStore interface {
	Create(ctx context.Context, s *domain.Session) error
	Get(ctx context.Context, sessionID string) (*domain.Session, error)
	Save(ctx context.Context, s *domain.Session) error
}

// GoalService resolves the learner's active goal so the diagnostic can target
// the right domain concepts.
type GoalService interface {
	ActiveStructure(ctx context.Context, learnerID string) (goalText, structureID string, err error)
}

// StructureResolver maps a structure uuid to its domain slug.
type StructureResolver interface {
	StructureDomainSlug(ctx context.Context, structureID string) (string, error)
}

// GraphService reads ordered concept nodes for a domain and retrieves RAG context.
type GraphService interface {
	TopoConcepts(ctx context.Context, domainSlug string) ([]NodeRef, error)
	GetResources(domainSlug, nodeID string) []string
}

type NodeRef struct {
	NodeID string
	Name   string
}

type ProfileService interface {
	GetPriorExperience(ctx context.Context, learnerID string) (string, error)
	GetRole(ctx context.Context, learnerID string) (string, error)
}

type QuestionData struct {
	Prompt        string   `json:"prompt"`
	Options       []string `json:"options"`
	CorrectOption int      `json:"correct_option"` // 0-based index
}

type LLMService interface {
	GenerateQuestionPrompt(ctx context.Context, role, priorLevel, conceptName, ragContext string) (*QuestionData, error)
	GenerateWeakAreasExplanation(ctx context.Context, role, priorLevel string, gaps []string, ragContext string) (string, error)
}

type CompetencyService interface {
	SaveBaseline(ctx context.Context, learnerID, nodeID, state string) error
}
