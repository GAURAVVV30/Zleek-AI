package domain

import "errors"

var (
	ErrSessionNotFound = errors.New("diagnostic session not found")
	ErrSessionComplete = errors.New("diagnostic session already completed")
	ErrNoActiveGoal    = errors.New("no active goal for diagnostic")
	ErrInvalidAnswer   = errors.New("invalid diagnostic answer")
)
