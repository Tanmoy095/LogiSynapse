// services/authentication-service/internal/domain/tenant/tenant.domain.go
package tenant

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
)

type Tenant struct {
	TenantID     uuid.UUID
	TenantName   string
	TenantStatus TenantStatus
	OwnerUserID  uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// EffectiveRole determines what permissions a user has.
// This is the core logic that fixes the dual-source-of-truth.
func (t *Tenant) GetEffectiveRole(userID uuid.UUID, membershipRole string) string {
	if t.OwnerUserID == userID {
		return "owner" // Owner wins, regardless of what's in the membership table.
	}
	return membershipRole
}
