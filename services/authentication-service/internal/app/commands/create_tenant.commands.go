// services/authentication-service/internal/app/commands/create_tenant.commands.go
package commands

import (
	"context"
	"time"

	domainError "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
)

//Any active, authenticated user may create a tenant,
// and that user becomes the owner of that tenant.
//An authenticated, active user may create a tenant for themselves.
//The creator becomes the owner.

type CreateTenantHandler struct {
	// Implementation pending
	userRepo   repository.UserStore
	tenantRepo repository.TenantStore
}

func NewCreateTenantHandler(
	userRepo repository.UserStore,
	tenantRepo repository.TenantStore,
) *CreateTenantHandler {
	return &CreateTenantHandler{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

type CreateTenantParams struct {
	TenantName  string
	ActorUserID uuid.UUID // who is calling (from JWT)
	OwnerUserID uuid.UUID // who will own the tenant (usually same)
}

func (h *CreateTenantHandler) Handle(ctx context.Context, params CreateTenantParams) (uuid.UUID, error) {
	//Authorization: actor can only create tenant for themselves (for now)
	if params.ActorUserID != params.OwnerUserID {
		return uuid.Nil, domainError.ErrInvalidInput
	}
	//  Verify User Exists (Referential Integrity)
	// We assume the ID comes from a valid JWT, but the user might have been deleted 1ms ago.
	owner, err := h.userRepo.GetUserByID(ctx, params.OwnerUserID)
	if err != nil {
		return uuid.Nil, domainError.ErrUserNotFound
	}
	if owner.Status != "active" {
		return uuid.Nil, domainError.ErrUserSuspended
	}
	//Construct Tenant Entity
	newTenant := &tenant.Tenant{
		TenantID:     uuid.New(),
		TenantName:   params.TenantName,
		TenantStatus: tenant.TenantStatusActive,
		OwnerUserID:  params.OwnerUserID, // Single Source of Truth
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	//Persist Tenant In DB
	if err := h.tenantRepo.CreateTenantWithOwnership(ctx, newTenant); err != nil {
		return uuid.Nil, err
	}
	//  (Future) Publish "TenantCreated" event for Billing Service via Kafka

	return newTenant.TenantID, nil
}
