// services/authentication-service/internal/ports/repository/user_Store.go
package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/user"
	"github.com/google/uuid"
)

type UserStore interface {
	CreateUser(ctx context.Context, user *user.User) error
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status user.UserStatus) error
	SetPasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error
}

// Key rules:
// context.Context always first
// No sql.ErrNoRows leaks → return domain errors
// No optional params → explicit methods
