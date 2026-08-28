package domain

import "errors"

var (
	ErrProfileNotFound = errors.New("learner profile not found")
)

type LearnerProfile struct {
	LearnerID   string
	Preferences map[string]interface{}
}
