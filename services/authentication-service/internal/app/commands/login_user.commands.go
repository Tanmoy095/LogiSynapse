// services/authentication-service/internal/app/commands/login_user.commands.go
package commands

import (
	"context"
	"fmt"
	"time"

	domainError "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/session"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/crypto"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
)

type LoginUserHandler struct {
	userRepo     repository.UserStore
	tokenRepo    repository.RefreshTokenStore
	passwordHash crypto.PasswordHasher
	tokenSigner  crypto.TokenSigner
}

func NewLoginUserHandler(
	userRepo repository.UserStore,
	tokenRepo repository.RefreshTokenStore,
	passwordHash crypto.PasswordHasher,
	tokenSigner crypto.TokenSigner,
) *LoginUserHandler {
	return &LoginUserHandler{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		passwordHash: passwordHash,
		tokenSigner:  tokenSigner,
	}
}

type LoginParams struct {
	Email             string
	Password          string
	DeviceFingerprint string
	IPAddress         string
}
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	TokenType    string
}

func (h *LoginUserHandler) Handler(ctx context.Context, params LoginParams) (*LoginResult, error) {

	//Fetch the USer by Email
	user, err := h.userRepo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		// If DB fails, return internal error.
		// If UserNotFound, we MUST NOT return immediately if we want to be paranoid about timing attacks,
		// but typically we let the ErrorMapper handle the status code.
		// For high security, we would "fake" a password verify here to equalize timing.
		return nil, domainError.ErrInvalidCredentials
	}
	//Verify Password
	match, err := h.passwordHash.VerifyPassword(ctx, params.Password, user.PasswordHash)
	if err != nil || !match {
		return nil, domainError.ErrInvalidCredentials
	}
	//Check Invariants
	if user.Status == "suspended" {
		return nil, domainError.ErrUserSuspended
	}
	if user.Status == "deleted" {
		return nil, domainError.ErrUserDeleted
	}
	//Generate Tokens
	// Access Token (JWT - Stateless)
	accessToken, jwtDuration, err := h.tokenSigner.SignAccessToken(ctx, crypto.AccessClaims{
		UserID:       user.UserID,
		UserEmail:    user.UserEmail,
		IsSuperAdmin: user.IsSuperAdmin,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}
	//Refresh Token (Opaque - Stateful)
	//WE Create the Domain Entity First
	FamilyID := uuid.New()                 // New login = New Family
	refreshTokenStr := uuid.New().String() //Opaque Token String. It needs to be long and random enough.
	// We must hash the refresh token before storage
	refreshTokenHash, err := h.passwordHash.HashPassword(ctx, refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 Days
	refreshTokenEntity := &session.RefreshToken{
		TokenID:        uuid.New(),
		UserID:         user.UserID,
		TokenHash:      refreshTokenHash,
		FamilyID:       FamilyID,
		IssuedAt:       time.Now(),
		ExpiresAt:      expiresAt,
		DeviceMetadata: params.DeviceFingerprint, // Storing Device Fingerprint
		// ReplacedBy and RevokedAt are nil for new sessions.
	}

	//Persist Session IN DB
	if err := h.tokenRepo.CreateRefreshToken(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}
	//return Result
	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,              // Return the RAW string to user, we stored the HASH
		ExpiresIn:    int64(jwtDuration.Seconds()), // In seconds
		TokenType:    "Bearer",                     // Standard OAuth2 Token Type
	}, nil

}
