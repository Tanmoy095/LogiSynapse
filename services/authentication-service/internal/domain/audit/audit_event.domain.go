// services/authentication-service/internal/domain/audit/audit_event.domain.go
package audit

import (
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents an immutable security record.
// It answers: Who did what, where, when, and with what context?
type AuditEvent struct {
	ID          uuid.UUID
	ActorUserID *uuid.UUID // Pointer allows for "System" actions or deleted users (SQL: SET NULL)
	TenantID    *uuid.UUID // Pointer allows for cross-tenant or system-wide events
	Action      string     // e.g., "MEMBER_INVITED", "TENANT_SUSPENDED"
	TargetID    *uuid.UUID // ID of the resource being acted upon
	IPAddress   string     // stored as INET in Postgres, string in Domain
	Metadata    map[string]any
	CreatedAt   time.Time
}

/*Audit logs are:
write-only, append-only immutable

 The Golden Audit Rule
Every successful state-changing command that affects security, authority, or ownership MUST emit exactly one audit event.
Not:
controllers
repositories
domain entities

Application Commands emit audit events
This keeps:
domain pure
infra dumb
audit complete

Where to emit audit events (exact places)
🔴 NEVER audit:
Reads
Failed authorization
Validation errors
✅ ALWAYS audit:
Tenant approved
Tenant rejected/request
Tenant created (platform)
Ownership transferred
Member invited / revoked/accept
User login (success)
User registration (optional, but common)


Super admin actions
*/
