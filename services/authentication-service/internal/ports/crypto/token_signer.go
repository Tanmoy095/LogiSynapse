package crypto

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// services/authentication-service/internal/ports/crypto/token_signer.go

// AccessClaims defines the standard data we bake into JWT Access Tokens
type AccessClaims struct {
	UserID       uuid.UUID
	UserEmail    string
	IsSuperAdmin bool
	Role         string // "admin" or "member"

}

// TokenSigner defines how we mint tokens.
// Implementation (Day 4) will use the crypto library.
type TokenSigner interface {
	// SignAccessToken generates a short-lived stateless JWT.
	SignAccessToken(ctx context.Context, claims AccessClaims) (token string, expiresIn time.Duration, err error)
}
