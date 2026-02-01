// services/authentication-service/internal/app/commands/create_tenant.commands.go
package commands

import (
	"context"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/audit"
	domainError "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
)

/*
CREATE TENANT — PLATFORM COMMAND

Golden Rules enforced:

Only SUPER ADMINS can create tenant by 2 rules

Rule A-->User → RequestTenantCmd .Super Admin → Approve / Reject
Owner is automatically the requester.


Rule B-->.without any user request. Platform Provisioning.. In this we will implement rule B
in this method
No user request needed
Super Admin → CreateTenantCmd
Owner and tenant name are explicitly chosen by super admin.

3. No membership is created for the owner
*/

type CreateTenantCmdByPlatform struct {
	// Implementation pending
	userRepo   repository.UserStore
	tenantRepo repository.TenantStore
	auditRepo  repository.AuditStore
}

func NewCreateTenantCmdByPlatform(
	userRepo repository.UserStore,
	tenantRepo repository.TenantStore,
	auditRepo repository.AuditStore,
) *CreateTenantCmdByPlatform {
	return &CreateTenantCmdByPlatform{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		auditRepo:  auditRepo,
	}
}

type CreateTenantParams struct {
	TenantName        string
	ActorUserID       uuid.UUID // who is executing this command (from JWT)
	IsActorSuperAdmin bool      // Is the actor a Super Admin? Injected by auth middleware
	OwnerUserID       uuid.UUID // Who will own the tenant
}

func (h *CreateTenantCmdByPlatform) Handle(ctx context.Context, params CreateTenantParams) (uuid.UUID, error) {
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

	// 🔐 AUDIT EVENT
	event := &audit.AuditEvent{
		ID:          uuid.New(),
		ActorUserID: &params.ActorUserID,
		TenantID:    &newTenant.TenantID,
		Action:      "TENANT_CREATED_BY_PLATFORM",
		TargetID:    &newTenant.TenantID,
		Metadata: map[string]any{
			"tenant_name": newTenant.TenantName,
			"owner_id":    newTenant.OwnerUserID,
		},
		CreatedAt: time.Now().UTC(),
	}

	_ = h.auditRepo.Append(ctx, event) // audit failure should NOT block creation

	return newTenant.TenantID, nil
}
