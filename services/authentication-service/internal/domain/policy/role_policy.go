// services/authentication-service/internal/domain/policy/role_policy.go
package policy

import (
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"
	"github.com/google/uuid"
)

// EvaluateEffectiveRole calculates the final authority a user has.
// This is the SINGLE SOURCE OF TRUTH for permissions.
func EvaluateEffectiveRole(tenantOwnerID, userID uuid.UUID, assignedRole membership.MemberShipRole) membership.MemberShipRole {
	// Invariant: The user defined as the owner in the Tenant table
	// ALWAYS gets RoleOwner, overriding the membership table.
	if tenantOwnerID == userID {
		return membership.MemberShipRoleOwner
	}

	// Fallback to the role stored in the membership table (Admin/Member).
	return assignedRole
}
