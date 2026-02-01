// services/authentication-service/internal/app/commands/register_User.commands.go
package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/audit"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/user"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/crypto"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
)

type RegisterUserHandler struct {
	UserRepo     repository.UserStore
	passwordHash crypto.PasswordHasher
	auditRepo    repository.AuditStore
}

func NewRegisterUserHandler(
	userRepo repository.UserStore,
	passwordHash crypto.PasswordHasher,
	auditRepo repository.AuditStore,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		UserRepo:     userRepo,
		passwordHash: passwordHash,
		auditRepo:    auditRepo,
	}
}

type RegisterUserParams struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

func (h *RegisterUserHandler) Handle(ctx context.Context, params RegisterUserParams) (uuid.UUID, error) {
	//  Invariant: Password Complexity
	// (Usually done in Proto validation, but good to have a domain check if logic is complex)

	//  Hash Password (CPU Intensive)
	// We do this BEFORE any DB transaction to keep transactions short.
	passwordHash, err := h.passwordHash.HashPassword(ctx, params.Password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to hash password: %w", err)
	}
	//Construct Entity
	newUser := &user.User{
		UserID:       uuid.New(),
		UserEmail:    params.Email,
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		PasswordHash: passwordHash,
		Status:       user.UserStatusActive,
		IsSuperAdmin: false, // Registration NEVER creates a super admin
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	//persist User In DB
	// We rely on the Repository to return ErrEmailAlreadyExists if the unique constraint violates.
	if err := h.UserRepo.CreateUser(ctx, newUser); err != nil {
		return uuid.Nil, err // Returns mapped domain error
	}
	// 🔐 AUDIT EVENT
	event := &audit.AuditEvent{
		ID:          uuid.New(),
		ActorUserID: &newUser.UserID,
		Action:      "USER_REGISTERED",
		CreatedAt:   time.Now().UTC(),
	}

	_ = h.auditRepo.Append(ctx, event)

	return newUser.UserID, nil
}
