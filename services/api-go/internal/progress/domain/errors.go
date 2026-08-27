package domain

import "errors"

var (
	ErrInvalidEvent   = errors.New("invalid engagement event")
	ErrEvidenceFailed = errors.New("failed to record evidence")
)
