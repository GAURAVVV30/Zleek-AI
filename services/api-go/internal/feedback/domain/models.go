package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidFeedback = errors.New("invalid feedback")
)

type FeedbackTargetType string

const (
	TargetResource FeedbackTargetType = "resource"
	TargetPath     FeedbackTargetType = "path_decision"
)

type FeedbackRecord struct {
	ID         string
	LearnerID  string
	TargetType FeedbackTargetType
	TargetID   string
	Rating     float64
	Comment    string
	CreatedAt  time.Time
}
