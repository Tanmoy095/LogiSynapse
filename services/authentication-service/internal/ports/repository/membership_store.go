// services/authentication-service/internal/ports/repository/membership_store.go

package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"
	"github.com/google/uuid"
)

type MemberShipStore interface {
	// Define methods for membership store here
	CreateMembership(ctx context.Context, membership *membership.MemberShip) error
	UpdateMembershipStatus(ctx context.Context, membership *membership.MemberShip) error
	GetMembersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]membership.MemberShip, error)
	ListMembersByUserID(ctx context.Context, userID uuid.UUID) ([]*membership.MemberShip, error)
	GetMember(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*membership.MemberShip, error)
	UpdateMemberRole(ctx context.Context, userID, tenantID uuid.UUID, role membership.Role) error
	// UpsertMembership is CRITICAL for Step 2.
	// Logic: If (user_id, tenant_id) exists, update Role/Status. If not, Insert.
	// Why? The old owner might already have a membership row (ignored by policy) or might not.
	// We need to force them to be an Admin in one atomic call.
	UpsertMembership(ctx context.Context, membership *membership.MemberShip) error
}
