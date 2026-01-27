//services/authentication-service/internal/domain/user/user.domain.go

package user

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusDeleted   UserStatus = "deleted"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	UserID       uuid.UUID
	UserEmail    string
	FirstName    string
	LastName     string
	PasswordHash string // Empty if OAuth user
	Status       UserStatus
	IsSuperAdmin bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
