package models

import (
	"time"

	"github.com/google/uuid"
)

// LoginAttempt tracks failed login attempts
type LoginAttempt struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Email      string    `json:"email" db:"email"`
	IPAddress  string    `json:"ipAddress" db:"ip_address"`
	Successful bool      `json:"successful" db:"successful"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

// AccountLockout represents a locked account
type AccountLockout struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"userId" db:"user_id"`
	LockedAt    time.Time  `json:"lockedAt" db:"locked_at"`
	LockedUntil time.Time  `json:"lockedUntil" db:"locked_until"`
	Reason      string     `json:"reason" db:"reason"`
	UnlockedAt  *time.Time `json:"unlockedAt,omitempty" db:"unlocked_at"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
}

// Role represents a user role for RBAC
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Permissions []string  `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// UserRole represents the mapping between users and roles
type UserRole struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	RoleID    uuid.UUID `json:"roleId" db:"role_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}
