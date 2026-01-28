// services/authentication-service/internal/domain/policy/role_policy.go
package policy

import (
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"
	"github.com/google/uuid"
)

// EffectiveRole calculates the actual authority a user has within a tenant.
// RULE: Tenant Ownership (in tenants table) overrides any specific role
// assigned in the memberships table. This prevents "dual source of truth" drift.
func EvaluateEffectiveRole(tenantOwnerID, userID uuid.UUID, assignedRole membership.MemberShipRole) membership.MemberShipRole {
	// Invariant: The user defined as the owner in the Tenant table
	// ALWAYS gets RoleOwner, overriding the membership table.
	if tenantOwnerID == userID {
		return membership.MemberShipRoleOwner
	}

	// Fallback to the role stored in the membership table (Admin/Member).
	return assignedRole
}

/*This is directionally correct, but not senior-clean.
Why?
Domain logic should not accept raw strings
This leaks infrastructure assumptions into the domain
It allows invalid states (membershipRole = "banana")
Interview-grade correction (you will do this later)
Accept membership.MemberShipRole
Return a strong enum, not string
Ownership logic should live in policy, not entity
We’ll fix this on Day 6 (Policy Layer).
For now: acceptable, but I see it.*/
