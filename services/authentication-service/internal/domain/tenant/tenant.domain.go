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

// TransferOwnership handles the domain state change.
//It does not handle a  database persistence or side effects.(like demoting the old owner).

func (t *Tenant) TransferOwnership(newOwnerUserID uuid.UUID) {
	//Invarient : ownership cannot be transfered to the same user and nil user.
	if newOwnerUserID == uuid.Nil || newOwnerUserID == t.OwnerUserID {
		return
	}
	t.OwnerUserID = newOwnerUserID
	t.UpdatedAt = time.Now().UTC()
}
