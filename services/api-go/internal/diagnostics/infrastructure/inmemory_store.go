package infrastructure

import (
	"context"
	"sync"

	"github.com/hcl-backend/services/api-go/internal/diagnostics/domain"
)

// InMemorySessionStore keeps diagnostic sessions for the lifetime of the Go
// process. Sessions are short-lived (a single onboarding quiz) so persistence
// is not required; the interface allows swapping a durable store later.
type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{sessions: map[string]*domain.Session{}}
}

func (s *InMemorySessionStore) Create(ctx context.Context, session *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.SessionID] = session
	return nil
}

func (s *InMemorySessionStore) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	cloned := *session
	cloned.Answers = make(map[string]string, len(session.Answers))
	for k, v := range session.Answers {
		cloned.Answers[k] = v
	}
	cloned.Prompts = make(map[string]string, len(session.Prompts))
	for k, v := range session.Prompts {
		cloned.Prompts[k] = v
	}
	cloned.CorrectAnswers = make(map[string]string, len(session.CorrectAnswers))
	for k, v := range session.CorrectAnswers {
		cloned.CorrectAnswers[k] = v
	}
	cloned.Questions = make([]domain.Question, len(session.Questions))
	copy(cloned.Questions, session.Questions)
	return &cloned, nil
}

func (s *InMemorySessionStore) Save(ctx context.Context, session *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.SessionID] = session
	return nil
}
