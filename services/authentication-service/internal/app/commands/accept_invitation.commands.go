// internal/app/commands/accept_invitation.go
package commands

import (
	"context"
	"time"

	"github.com/google/uuid"

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
}

func NewAcceptInvitationCmd(r repository.MemberShipStore) *AcceptInvitationCmd {
	return &AcceptInvitationCmd{membershipRepo: r}
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
		return c.membershipRepo.UpdateMembershipStatus(ctx, member)

	default:
		return domainErr.ErrInvalidState
	}
}
