//services/authentication-service/internal/ports/repository/tenant_store.go

package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/google/uuid"
)

type TenantStore interface {
	// CreateTenantWithOwnership ensures the Tenant and the Owner's Membership
	// are created in a single Database Transaction.
	CreateTenantWithOwnership(ctx context.Context, t *tenant.Tenant) error
	GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*tenant.Tenant, error)
	//suspended or active status update
	UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status tenant.TenantStatus) error
	ListTenantsByOwnerID(ctx context.Context, ownerUserID uuid.UUID) ([]tenant.Tenant, error)
	// UpdateTenant persists changes to the tenant (e.g., owner_user_id, )
	UpdateTenant(ctx context.Context, tenant *tenant.Tenant) error
}
