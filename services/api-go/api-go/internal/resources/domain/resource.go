package domain

import (
	"time"
)

type ResourceStatus string

const (
	StatusCandidate ResourceStatus = "candidate"
	StatusPublished ResourceStatus = "published"
	StatusRetired   ResourceStatus = "retired"
	StatusFlagged   ResourceStatus = "flagged"
)

type FreshnessStatus string

const (
	FreshnessFresh      FreshnessStatus = "fresh"
	FreshnessStale      FreshnessStatus = "stale"
	FreshnessUnverified FreshnessStatus = "unverified"
)

type Resource struct {
	ID              string
	URL             string
	Source          *string
	Author          *string
	ResourceType    string
	Difficulty      *string
	AuthorityScore  *float64
	ProvenanceNote  *string
	Status          ResourceStatus
	LastCheckedAt   *time.Time
	FreshnessStatus FreshnessStatus
	CuratedBy       *string
	CuratedAt       *time.Time
	CreatedAt       time.Time
}

func (r *Resource) Publish(curatorID string) error {
	if r.Status == StatusPublished {
		return ErrInvalidStateTransition
	}
	r.Status = StatusPublished
	r.CuratedBy = &curatorID
	now := time.Now()
	r.CuratedAt = &now
	return nil
}

func (r *Resource) Retire() error {
	if r.Status == StatusRetired {
		return ErrInvalidStateTransition
	}
	r.Status = StatusRetired
	return nil
}

type ResourceConcept struct {
	ResourceID    string
	ConceptID     string
	RelevanceNote *string
}

type ResourceQualitySignal struct {
	ResourceID         string
	AvgRating          *float64
	FeedbackCount      int
	OutcomeCorrelation *float64
	UpdatedAt          time.Time
}
