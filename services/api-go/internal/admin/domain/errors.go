package domain

import "errors"

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidRole     = errors.New("invalid role")
	ErrInvalidStatus   = errors.New("invalid status")
	ErrSelfDemotion    = errors.New("cannot demote or suspend self")
)
