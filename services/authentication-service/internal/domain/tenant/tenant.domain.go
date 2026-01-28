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
