//services/authentication-service/internal/ports/crypto/hasher.crypto.go

package crypto

import "context"

// PasswordHasher defines the contract for password security.
// We use an interface so the Domain doesn't care about the algorithm.
type PasswordHasher interface {
	HashPassword(ctx context.Context, password string) (string, error)
	VerifyPassword(ctx context.Context, password, encodedHash string) (bool, error)
}
