// internal/app/commands/accept_invitation.go
package commands

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/audit"
	domainErr "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
)

/*
ACCEPT INVITATION — STATE TRANSITION

Golden Rules enforced:
1. Pending → Active only
2. Revoked invites cannot be accepted
3. Idempotent behavior
*/

type AcceptInvitationCmd struct {
	membershipRepo repository.MemberShipStore
	auditRepo      repository.AuditStore
}

func NewAcceptInvitationCmd(r repository.MemberShipStore, a repository.AuditStore) *AcceptInvitationCmd {
	return &AcceptInvitationCmd{membershipRepo: r, auditRepo: a}
}

func (c *AcceptInvitationCmd) Handle(ctx context.Context, userID, tenantID uuid.UUID) error {
	member, err := c.membershipRepo.GetMember(ctx, userID, tenantID)
	if err != nil || member == nil {
		return domainErr.ErrMembershipNotFound
	}

	switch member.MemberShipStatus {
	case membership.StatusActive:
		// Idempotent success
		return nil

	case membership.StatusRevoked:
		return domainErr.ErrUnauthorized

	case membership.StatusPending:
		member.MemberShipStatus = membership.StatusActive
		member.UpdatedAt = time.Now().UTC()

		if err := c.membershipRepo.UpdateMembershipStatus(ctx, member); err != nil {
			return err
		}

		// 🔐 AUDIT EVENT
		event := &audit.AuditEvent{
			ID:          uuid.New(),
			ActorUserID: &userID,
			TenantID:    &tenantID,
			Action:      "MEMBERSHIP_ACCEPTED",
			TargetID:    &userID,
			Metadata: map[string]any{
				"role": member.MemberShipRole,
			},
			CreatedAt: time.Now().UTC(),
		}

		_ = c.auditRepo.Append(ctx, event)
		return nil

	default:
		return domainErr.ErrInvalidState
	}
	// FUTURE:
	// - Emit MembershipActivated event

}
