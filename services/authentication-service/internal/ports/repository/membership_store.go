// services/authentication-service/internal/ports/repository/membership_store.go

package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"
	"github.com/google/uuid"
)

type MemberShipStore interface {
	// Define methods for membership store here
	AddMember(ctx context.Context, membership *membership.MemberShip) error
	GetMembersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]membership.MemberShip, error)
	ListMembersByUserID(ctx context.Context, userID uuid.UUID) ([]*membership.MemberShip, error)
	GetMember(ctx context.Context, userID, tenantID uuid.UUID) (*membership.MemberShip, error)
	UpdateMemberRole(ctx context.Context, userID, tenantID uuid.UUID, role membership.MemberShipRole) error
}
