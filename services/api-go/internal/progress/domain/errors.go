package domain

import "errors"

var (
	ErrInvalidEvent       = errors.New("invalid engagement event")
	ErrEvidenceFailed     = errors.New("failed to record evidence")
	ErrNoActivePath       = errors.New("no active learning path")
	ErrPrerequisiteNotMet = errors.New("previous module must be completed first")
)
