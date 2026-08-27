package domain

import (
	"errors"
)

var (
	ErrActivePathNotFound      = errors.New("active path not found")
	ErrNoActiveGoal            = errors.New("no active goal found")
	ErrInvalidAIProposal       = errors.New("invalid AI proposal")
	ErrInvalidPrerequisite     = errors.New("invalid prerequisite ordering in proposal")
	ErrUnknownConcept          = errors.New("proposal contains unknown concept")
	ErrUnknownResource         = errors.New("proposal contains unknown resource")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrConcurrentUpdate        = errors.New("concurrent update conflict")
	ErrAIUnavailable           = errors.New("AI service unavailable")
)
