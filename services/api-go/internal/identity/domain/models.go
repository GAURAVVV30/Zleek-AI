package domain

import (
	"time"
)

type UserRole string

const (
	RoleLearner UserRole = "learner"
	RoleCurator UserRole = "curator"
	RoleAdmin   UserRole = "admin"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         UserRole
	Status       UserStatus
	FullName     string
	Timezone     string
	Theme        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
