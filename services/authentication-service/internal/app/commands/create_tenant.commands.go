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

/*
CREATE TENANT — PLATFORM COMMAND

Golden Rules enforced:
1. Only SUPER ADMINS can create tenants
2. Ownership is a PROPERTY of the tenant
3. No membership is created for the owner
*/

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
	TenantName        string
	ActorUserID       uuid.UUID // who is executing this command (from JWT)
	IsActorSuperAdmin bool      // Is the actor a Super Admin? Injected by auth middleware
	OwnerUserID       uuid.UUID // Who will own the tenant
}

func (h *CreateTenantHandler) Handle(ctx context.Context, params CreateTenantParams) (uuid.UUID, error) {
	// 1. Enforce Rule 1: Platform Authority
	// "Only SUPER ADMINS can create tenants"
	if !params.IsActorSuperAdmin {
		return uuid.Nil, domainError.ErrUnauthorized // "You are not a god"
	}

	// Referential integrity: owner must exist & be active
	owner, err := h.userRepo.GetUserByID(ctx, params.OwnerUserID)
	if err != nil {
		return uuid.Nil, domainError.ErrUserNotFound
	}
	if owner.Status != "active" {
		return uuid.Nil, domainError.ErrUserSuspended
	}
	// RULE 2 — OWNER IS A TENANT PROPERTY
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
	// (Future) Emit TenantCreated event

	return newTenant.TenantID, nil
}
