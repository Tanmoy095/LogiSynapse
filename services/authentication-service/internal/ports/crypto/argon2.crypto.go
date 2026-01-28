//services/authentication-service/internal/ports/crypto/argon2.crypto.go

package crypto

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Params defines the memory and CPU cost factors for Argon2id.
type Params struct {
	Memory      uint32 // RAM usage in KB (e.g., 64*1024 = 64MB)
	Iterations  uint32 // Number of passes over the memory
	Parallelism uint8  // Number of threads/cores to use
	SaltLength  uint32 // Random salt length in bytes
	KeyLength   uint32 // Final hash length in bytes
}

// DefaultParams are tuned for 2026 security standards.
// Balanced for a typical cloud container (0.5 - 1 CPU core).
var DefaultParams = &Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// argon2Hasher is the private implementation of the PasswordHasher interface.
type argon2Hasher struct {
	params *Params
}

// NewArgon2Hasher is the Constructor (Factory).
// It returns the interface, keeping the implementation details private.
func NewArgon2Hasher(p *Params) PasswordHasher {
	if p == nil {
		p = DefaultParams
	}
	return &argon2Hasher{params: p}
}

// HashPassword implements the PasswordHasher interface.
func (h *argon2Hasher) HashPassword(ctx context.Context, password string) (string, error) {
	// 1. Generate a cryptographically secure random salt.
	// We do this EVERY time so identical passwords have different hashes.
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("crypto/rand failed: %w", err)
	}

	// 2. Derive the key using Argon2id (Memory-hard against GPU/ASIC attacks).
	// Note: We ignore ctx here as argon2 doesn't support cancellation,
	// but we keep it to satisfy the interface.
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	// 3. Encode to PHC format string.
	// This format includes the params so we can verify the hash even if we change defaults later.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.Memory, h.params.Iterations, h.params.Parallelism, b64Salt, b64Hash,
	)

	return encoded, nil
}

// VerifyPassword implements the PasswordHasher interface.
func (h *argon2Hasher) VerifyPassword(ctx context.Context, password, encodedHash string) (bool, error) {
	// 1. Parse the existing hash to get the salt and parameters used at creation time.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	// 2. Hash the input password using the extracted salt and params.
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)

	// 3. Constant-time comparison to prevent side-channel (timing) attacks.
	// subtle.ConstantTimeCompare prevents attackers from guessing the hash based on response time.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// decodeHash is a helper to parse the "$argon2id$v=19$m=65536..." string.
func decodeHash(encodedHash string) (p *Params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, fmt.Errorf("hash has wrong parts")
	}

	var version int
	if _, err = fmt.Sscanf(vals[2], "v=%d", &version); err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible version")
	}

	p = &Params{}
	if _, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism); err != nil {
		return nil, nil, nil, err
	}

	if salt, err = base64.RawStdEncoding.DecodeString(vals[4]); err != nil {
		return nil, nil, nil, err
	}

	if hash, err = base64.RawStdEncoding.DecodeString(vals[5]); err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
