package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/audit"
	domainErr "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/policy"
	"golang.org/x/sync/errgroup"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
)

type RevokeMembershipCmd struct {
	membershipRepo repository.MemberShipStore
	tenantRepo     repository.TenantStore
	auditRepo      repository.AuditStore
}

func NewRevokeMembershipCmd(
	membershipRepo repository.MemberShipStore,
	tenantRepo repository.TenantStore,
	auditRepo repository.AuditStore,
) *RevokeMembershipCmd {
	return &RevokeMembershipCmd{
		membershipRepo: membershipRepo,
		tenantRepo:     tenantRepo,
		auditRepo:      auditRepo,
	}
}

type RevokeMembershipParams struct {
	TenantID    uuid.UUID
	ActorUserID uuid.UUID // Who is doing the firing?

	TargetUserID uuid.UUID // Who is getting fired?
}

func (cmd *RevokeMembershipCmd) Handle(ctx context.Context, params RevokeMembershipParams) error {
	//Self-Protection
	// Revoking yourself is a different workflow ("Leave Tenant") with different checks (e.g., owner cannot leave).
	if params.ActorUserID == params.TargetUserID {
		return domainErr.ErrCannotRevokeSelf
	}
	// Fetch Tenant (Single Source of Truth for Ownership)
	// We need this first to calculate EffectiveRoles accurately.
	tenant, err := cmd.tenantRepo.GetTenantByID(ctx, params.TenantID)
	if err != nil {
		return domainErr.ErrTenantNotFound
	}
	//Invariant: Continuity Protection
	// "Owner cannot be revoked".
	// We check the TENANT property, not the membership role.
	if tenant.OwnerUserID == params.TargetUserID {
		return domainErr.ErrCannotRevokeOwner
	}
	// Fetch both memberships in parallel (Fan-out)
	// We need to know the rank of the Actor and the rank of the Target.
	var (
		actorMem  *membership.MemberShip
		targetMem *membership.MemberShip
	)
	g, ctx := errgroup.WithContext(ctx)
	// Task A: Get Actor
	g.Go(func() error {
		member, err := cmd.membershipRepo.GetMember(ctx, params.ActorUserID, params.TenantID)
		// If actor isn't in DB, they might be Owner (who doesn't require a membership row).
		// We handle nil in EffectiveRole, so we ignore Not Found errors here,
		// but we should report DB errors.
		if err != nil && err != domainErr.ErrMembershipNotFound {
			return err
		}
		actorMem = member
		return nil

	})
	// Task B: Get Target
	g.Go(func() error {
		m, err := cmd.membershipRepo.GetMember(ctx, params.TargetUserID, params.TenantID)
		if err != nil {
			// If target doesn't exist, we can't revoke them.
			if err == domainErr.ErrMembershipNotFound {
				return domainErr.ErrMembershipNotFound
			}
			return err
		}
		targetMem = m
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	// 5. Policy Engine: Calculate Ranks
	actorRole := policy.EffectiveRole(tenant.OwnerUserID, params.ActorUserID, actorMem)
	targetRole := policy.EffectiveRole(tenant.OwnerUserID, params.TargetUserID, targetMem)
	// Authorization Check: Can Actor revoke Target?
	// 6. Authorization Matrix (The Hierarchy Check)
	allowed := false

	switch actorRole {
	case membership.RoleOwner:
		// Owner can fire ANYONE (except self, checked above)
		allowed = true

	case membership.RoleAdmin:
		// Admin can ONLY fire Members.
		// Admin cannot fire Owner (checked above).
		// Admin cannot fire Admin (prevent coup).
		if targetRole == membership.RoleMember {
			allowed = true
		} else {
			return domainErr.ErrInsufficientPrivilege // "Cannot fire peer or superior"
		}

	case membership.RoleMember, membership.RoleNone:
		return domainErr.ErrUnauthorizedAction
	}

	if !allowed {
		return domainErr.ErrUnauthorizedAction
	}

	// 7. State Transition: Soft Delete
	// We do NOT delete the row. We set status to Revoked.
	// This preserves the "Audit Trail" that this user *was* here.
	if targetMem.MemberShipStatus == membership.StatusRevoked {
		return nil // Idempotent success
	}

	targetMem.MemberShipStatus = membership.StatusRevoked
	targetMem.UpdatedAt = time.Now().UTC()

	if err := cmd.membershipRepo.UpdateMembershipStatus(ctx, targetMem); err != nil {
		return fmt.Errorf("failed to revoke membership: %w", err)
	}

	// 8. Audit Log
	event := &audit.AuditEvent{
		ID:          uuid.New(),
		ActorUserID: &params.ActorUserID,
		TenantID:    &params.TenantID,
		Action:      "MEMBERSHIP_REVOKED",
		TargetID:    &params.TargetUserID,
		Metadata: map[string]any{
			"target_role_at_time": targetRole,
			"actor_role":          actorRole,
		},
		CreatedAt: time.Now().UTC(),
	}

	return cmd.auditRepo.Append(ctx, event)
}
