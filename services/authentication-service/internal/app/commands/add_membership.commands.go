// services/authentication-service/internal/app/commands/add_membership.commands.go
package commands

import (
	"context"
	"time"

	domainError "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/policy"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/user"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type AddMembershipCmd struct {
	// Implementation pending
	userRepo       repository.UserStore
	tenantRepo     repository.TenantStore
	membershipRepo repository.MemberShipStore
}

func NewAddMembershipCmd(
	u repository.UserStore,
	t repository.TenantStore,
	m repository.MemberShipStore,
) *AddMembershipCmd {
	return &AddMembershipCmd{u, t, m}
}

type AddMembershipParams struct {
	TenantID    uuid.UUID
	ActorUserID uuid.UUID // person performing the action (the "inviter") should be admin and Owner

	TargetUserEmail string          //Email of the person being invited
	Role            membership.Role // Role to assign: admin/member
}

func (h *AddMembershipCmd) Handle(ctx context.Context, params AddMembershipParams) error {
	// Implementation pending
	// Invariant: Cannot invite owner
	if params.Role == membership.RoleOwner {
		return domainError.ErrInvalidInput
	}

	//efficiency and safety database lookups at the exact same time instead of one after the other
	// this pattern called "Fan-out" to perform two database lookups concurrently.
	var (
		targetUser   *user.User
		targetTenant *tenant.Tenant
	) //will hold the data once it's fetched from the database

	//errGroup creates a group of "tasks" (goroutines). If any one of these tasks fails, the errgroup catches the error and can cancel the other tasks to save resources.
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		user, err := h.userRepo.GetUserByEmail(ctx, params.TargetUserEmail)
		if err != nil {
			return domainError.ErrUserNotFound
		}
		targetUser = user
		return nil
	})
	group.Go(func() error {
		tenant, err := h.tenantRepo.GetTenantByID(ctx, params.TenantID)
		if err != nil {
			return domainError.ErrTenantNotFound
		}
		targetTenant = tenant
		return nil
	})
	if err := group.Wait(); err != nil {
		return err
	}
	//The code pauses here. g.Wait() waits for both Task 1 and Task 2 to finish.

	//If both succeed: The code continues, and our variables (targetUser and targetTenant) are now full of data.

	//If any fail: It returns the error immediately, and the rest of the function stops.

	// task:-->// Find out what role the "Inviter/Actor" has in this tenant
	actorMembership, _ := h.membershipRepo.GetMember(ctx, params.ActorUserID, params.TenantID)
	// Determine the "Effective Role"
	effectiveRole := policy.EffectiveRole(targetTenant.OwnerUserID, params.ActorUserID, actorMembership)

	//only owner and admins can add members
	if effectiveRole != membership.RoleOwner && effectiveRole != membership.RoleAdmin {
		return domainError.ErrNotTenantAdmin
	}
	// Invariant: Cannot invite yourself
	if targetUser.UserID == params.ActorUserID {
		return domainError.ErrInvalidInput
	}

	// Idempotency: already invited or member
	existing, _ := h.membershipRepo.GetMember(ctx, targetUser.UserID, params.TenantID)
	if existing != nil {
		return domainError.ErrDuplicateMembership
	}
	// RULE 3 — INVITATION STARTS AS PENDING
	invite := &membership.MemberShip{
		UserID:           targetUser.UserID,
		TenantID:         params.TenantID,
		MemberShipRole:   params.Role,
		MemberShipStatus: membership.StatusPending, // Not active yet!
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	return h.membershipRepo.CreateMembership(ctx, invite)

}
