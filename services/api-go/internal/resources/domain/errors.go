package domain

import "errors"

var (
	ErrResourceNotFound       = errors.New("resource not found")
	ErrConceptNotFound        = errors.New("concept not found")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrDuplicateResource      = errors.New("resource URL already exists")
)
