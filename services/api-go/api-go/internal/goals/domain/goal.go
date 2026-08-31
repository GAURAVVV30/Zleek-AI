package domain

import (
	"time"
)

type GoalStatus string

const (
	StatusActive    GoalStatus = "active"
	StatusAchieved  GoalStatus = "achieved"
	StatusAbandoned GoalStatus = "abandoned"
)

type Goal struct {
	ID                   string
	LearnerID            string
	GoalText             string
	KnowledgeStructureID string
	Status               GoalStatus
	AchievedAt           *time.Time
	CreatedAt            time.Time
}
