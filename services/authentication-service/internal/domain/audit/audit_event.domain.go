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
