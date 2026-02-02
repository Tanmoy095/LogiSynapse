// services/authentication-service/internal/app/commands/transferTenant_ownership.commands.go
package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/audit"
	domainErr "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
)

/*
The "Continuity Problem": In a naive implementation, when you transfer ownership from Alice to Bob:

Bob becomes Owner.

Alice is removed from owner_user_id.

Result: Alice is now locked out of the company she just founded.

The Senior Solution (The "Safe Swap"): We perform an atomic Rotation:

Promote Bob: Update tenants.owner_user_id to Bob.

Demote Alice: Insert (or Update) a membership record for Alice with Role: Admin.

Reasoning: This ensures business continuity. The old owner steps down to Admin but doesn't lose access.

TRANSFER TENANT OWNERSHIP — STATE TRANSITION
Atomic Transaction: Swaps owner and creates admin membership in one go.

Actor Validation: Only Current Owner or Super Admin can do this.

Audit: Logs both the old and new owner IDs.
if the user is NOT the owner AND is NOT a super admin (both must be false to block)
Business Continuity: We didn't just change a database column; we thought about what happens to the human (the old owner) and ensured they remain an Admin.

Policy Leverage: We leveraged the EffectiveRole logic we wrote in Day 1. We realized we don't need to delete the new owner's old membership row because Tenant.OwnerUserID overrides it. This saves a database call.

Atomic Safety: If the audit log fails, or the downgrade fails, the ownership transfer rolls back. No "zombie state" where the tenant has a new owner but the log is missing.
*/
type TransTntOwnership struct {
	tenantRepo     repository.TenantStore
	membershipRepo repository.MemberShipStore
	userRepo       repository.UserStore
	auditRepo      repository.AuditStore
	txManager      repository.TransactionManager
}

func NewTransTntOwnership(t repository.TenantStore, m repository.MemberShipStore, u repository.UserStore, a repository.AuditStore, tx repository.TransactionManager) *TransTntOwnership {
	return &TransTntOwnership{
		tenantRepo:     t,
		membershipRepo: m,
		userRepo:       u,
		auditRepo:      a,
		txManager:      tx,
	}
}

type TransferOwnershipParams struct {
	tenantID     uuid.UUID
	ActorUserID  uuid.UUID
	IsSuperAdmin bool
	// We identify the new owner by ID.
	// In a UI, you might look them up by email first, but the command expects an ID.
	NewOwnerUserID uuid.UUID
}

func (cmd *TransTntOwnership) Handle(ctx context.Context, params TransferOwnershipParams) error {

	//validation
	if params.NewOwnerUserID == uuid.Nil {
		return domainErr.ErrInvalidInput
	}
	//  Transaction Scope
	return cmd.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		//Fetch Tenant (Locking row for update is ideal in SQL implementation)
		tnt, err := cmd.tenantRepo.GetTenantByID(txCtx, params.tenantID)
		if err != nil || tnt == nil {
			return domainErr.ErrTenantNotFound
		}
		//Authorization: Only Owner or Super Admin can transfer ownership
		if tnt.OwnerUserID != params.ActorUserID && !params.IsSuperAdmin {

			return domainErr.ErrUnauthorized
		}
		//New owner must be a valid user
		newOwner, err := cmd.userRepo.GetUserByID(txCtx, params.NewOwnerUserID)
		if err != nil || newOwner == nil {
			return domainErr.ErrUserNotFound
		}
		if newOwner.Status != "active" {
			return domainErr.ErrUserSuspended // Cannot transfer company to suspended user
		}
		// Capture Old Owner ID for processing
		oldOwnerID := tnt.OwnerUserID
		// Edge Case: Transferring to self? No-op.
		if oldOwnerID == params.NewOwnerUserID {
			return nil
		}
		//Domain Logic: Swap Ownership on Tenant Entity
		tnt.TransferOwnership(params.NewOwnerUserID)
		if err := cmd.tenantRepo.UpdateTenant(txCtx, tnt); err != nil {
			return fmt.Errorf("failed to update tenant owner: %w", err)
		}
		//  Continuity Logic: Demote Old Owner to Admin
		// We Upsert a membership for the old owner.
		oldOwnerMembership := &membership.MemberShip{
			UserID:           oldOwnerID,
			TenantID:         params.tenantID,
			MemberShipRole:   membership.RoleAdmin,    // Degrades to Admin
			MemberShipStatus: membership.StatusActive, // Ensures they can login
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}
		//UpsertMembership must handle "ON CONFLICT (user_id, tenant_id) DO UPDATE"
		if err := cmd.membershipRepo.UpsertMembership(ctx, oldOwnerMembership); err != nil {
			return fmt.Errorf("failed to downgrade old owner: %w", err)
		}
		// Note: We do NOT need to touch the New Owner's membership.
		// Even if they were a "Member" or "Pending", the "EffectiveRole" policy
		//  checks Tenant.OwnerUserID FIRST.
		// So they instantly become Owner in the eyes of the system.
		// 🔐 AUDIT EVENT
		event := &audit.AuditEvent{
			ID:          uuid.New(),
			ActorUserID: &params.ActorUserID,
			TenantID:    &params.tenantID,
			Action:      "TENANT_OWNERSHIP_TRANSFERRED",
			TargetID:    &params.NewOwnerUserID,
			Metadata: map[string]any{
				"old_owner_id": oldOwnerID,
				"new_owner_id": params.NewOwnerUserID,
				"initiated_by": "manual_transfer",
			},
			CreatedAt: time.Now().UTC(),
		}
		return cmd.auditRepo.Append(ctx, event)

	})
}
