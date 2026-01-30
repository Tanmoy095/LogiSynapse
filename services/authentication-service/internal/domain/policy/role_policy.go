// services/authentication-service/internal/domain/policy/role_policy.go
package policy

import (
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/membership"
	"github.com/google/uuid"
)

// EffectiveRole calculates the actual authority a user has.
//
// Golden Rules Enforced:
// 1. Tenant Owner is ALWAYS 'owner' (System of record: tenants table).
// 2. Pending invites grant NO authority (System of record: memberships table).
func EffectiveRole(tenantOwnerID, userID uuid.UUID, member *membership.MemberShip) membership.Role {
	// 1. Check Ownership (Absolute Truth)
	if tenantOwnerID == userID {
		return membership.RoleOwner
	}

	// 2. Check Membership Existence
	if member == nil {
		return membership.RoleNone
	}

	// 3. Check Invitation Status (Rule 3)
	// Even if DB says 'admin', if status is 'pending', they are nobody.
	if member.MemberShipStatus != membership.StatusActive {
		return membership.RoleNone
	}

	// 4. Return Stored Role ('admin' or 'member')
	return member.MemberShipRole
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
