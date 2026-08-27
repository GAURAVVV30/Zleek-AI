package domain

import "errors"

var (
	ErrGoalNotFound         = errors.New("goal not found")
	ErrAIProposalInvalid    = errors.New("ai proposal invalid")
	ErrKnowledgeUnpublished = errors.New("knowledge structure is not published")
	ErrDomainNotFound       = errors.New("domain not found")
)
