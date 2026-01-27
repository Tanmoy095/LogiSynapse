// services/authentication-service/internal/domain/session/session.domain.go
package session

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	TokenID  uuid.UUID
	UserId   uuid.UUID
	TenantId *uuid.UUID // Optional (nil if user not logged into specific tenant yet)
	// TokenHash is the SHA-256 hash of the raw, opaque refresh token string.
	// WHY: We never store raw tokens. If the database is leaked, hashes are useless to attackers.
	// HOW: When a user sends a token, we hash it in the app and query the DB for the match.
	TokenHash string

	// FamilyID groups all tokens generated from a single "Login Event."
	// WHY: To support "Token Rotation." If you log in on Chrome, that session is one family.
	// HOW: When a token is refreshed, the new token inherits this FamilyID. If we detect
	// a "Replay Attack" (using an old token), we revoke every token with this FamilyID.
	FamilyID uuid.UUID

	IssuedAt  time.Time
	ExpiresAt time.Time

	// RevokedAt is a "Kill Switch" timestamp.
	// WHY: Allows for instant session invalidation (Logout, Security Breach, or Admin Action).
	// HOW: If this is not nil, the token is dead. It is a soft-delete mechanism.
	RevokedAt *time.Time

	// ReplacedBy links this token to the next one in the rotation chain.
	// WHY: To detect Replay Attacks and maintain a "Searchable Audit Trail" of the session.
	// HOW: When Token A is used to get Token B, Token A.ReplacedBy becomes Token B.ID.
	// If a token has a ReplacedBy value but someone tries to use it again, we know it's a theft.
	ReplacedBy *uuid.UUID

	// DeviceMetadata stores a fingerprint of the client (IP, User-Agent, etc.).
	// WHY: For risk-based authentication and "Known Device" detection.
	// HOW: If a refresh token issued for "Chrome/Windows/London" suddenly appears
	// from "Firefox/Linux/unknown-location," we flag it as a high-risk event.
	DeviceMetadata string
}
