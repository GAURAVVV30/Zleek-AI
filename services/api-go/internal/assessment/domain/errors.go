package domain

import "errors"

var (
	ErrAssessmentNotFound = errors.New("assessment not found")
	ErrConceptNotFound    = errors.New("concept not found")
	ErrInvalidSubmission  = errors.New("invalid submission")
	ErrInvalidAIResult    = errors.New("invalid AI evaluation result")
	ErrAIUnavailable      = errors.New("ai evaluation service unavailable")
	ErrInvalidType        = errors.New("invalid assessment type")
)
