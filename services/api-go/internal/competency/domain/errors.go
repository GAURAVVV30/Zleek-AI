package domain

import "errors"

var (
	ErrCompetencyNotFound = errors.New("competency not found")
	ErrConceptNotFound    = errors.New("concept not found")
	ErrInvalidEvidence    = errors.New("invalid evidence")
	ErrInvalidState       = errors.New("invalid competency state")
)
