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
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
