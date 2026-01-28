// services/authentication-service/internal/domain/membership/membership.domain.go
package membership

import (
	"time"

	"github.com/google/uuid"
)

type MemberShipRole string

const (
	MemberShipRoleAdmin  MemberShipRole = "admin"
	MemberShipRoleMember MemberShipRole = "member"
	// RoleOwner is not assigned in the DB, but used in the Policy layer.
	MemberShipRoleOwner MemberShipRole = "owner"
)

type MemberShip struct {
	UserID         uuid.UUID
	TenantID       uuid.UUID
	MemberShipRole MemberShipRole
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
