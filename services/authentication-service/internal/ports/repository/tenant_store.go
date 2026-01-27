//services/authentication-service/internal/ports/repository/tenant_store.go

package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/google/uuid"
)

type TenantStore interface {
	CreateTenant(ctx context.Context, tenant *tenant.Tenant) error
	GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*tenant.Tenant, error)
	UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status tenant.TenantStatus) error
	GetTenantByOwnerID(ctx context.Context, ownerUserID uuid.UUID) ([]tenant.Tenant, error)
}
