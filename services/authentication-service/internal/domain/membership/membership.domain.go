// services/authentication-service/internal/domain/membership/membership.domain.go
package membership

import (
	"time"

	"github.com/google/uuid"
)

type Role string
type MemberShipStatus string

const (
	// Rule 2: 'Owner' is a computed concept, not a stored role.
	// We keep RoleOwner here for return values (Policy), but it never goes into DB.
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleNone   Role = "" // For unauthorized users
)
const (
	StatusPending MemberShipStatus = "pending" // Invite sent, not accepted
	StatusActive  MemberShipStatus = "active"  // Accepted
	StatusRevoked MemberShipStatus = "revoked" // Banished
)

type MemberShip struct {
	UserID           uuid.UUID
	TenantID         uuid.UUID
	MemberShipRole   Role
	MemberShipStatus MemberShipStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
