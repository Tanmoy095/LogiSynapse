//services/authentication-service/internal/ports/repository/refresh_token_store.go

package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/session"
	"github.com/google/uuid"
)

type RefreshTokenStore interface {
	CreateRefreshToken(ctx context.Context, token *session.RefreshToken) error
	GetTokenByHash(ctx context.Context, hash string) (*session.RefreshToken, error)
	// RevokeTokenFamily is crucial for the "Replay Detection" invariant
	RevokeTokenFamily(ctx context.Context, familyID uuid.UUID) error
	// RotateToken handles the atomic swap: mark old as replaced, insert new
	RotateToken(ctx context.Context, oldTokenID uuid.UUID, newToken *session.RefreshToken) error
}
