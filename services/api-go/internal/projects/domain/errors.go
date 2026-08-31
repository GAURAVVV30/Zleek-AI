package domain

import "errors"

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrInvalidArtifact     = errors.New("invalid artifact reference")
	ErrSubmissionFailed    = errors.New("submission failed")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidConcept      = errors.New("invalid concept")
	ErrNoProjectForConcept = errors.New("concept does not have a project assessment")
)
