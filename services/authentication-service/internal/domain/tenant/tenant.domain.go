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

/*This is directionally correct, but not senior-clean.
Why?
Domain logic should not accept raw strings
This leaks infrastructure assumptions into the domain
It allows invalid states (membershipRole = "banana")
Interview-grade correction (you will do this later)
Accept membership.MemberShipRole
Return a strong enum, not string
Ownership logic should live in policy, not entity
We’ll fix this on Day 6 (Policy Layer).
For now: acceptable, but I see it.*/
