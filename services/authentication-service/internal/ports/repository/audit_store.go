// services/authentication-service/internal/ports/repository/audit_store.go

package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/audit"
)

type AuditStore interface {
	// Append is "Write Only". We rarely update or delete audit logs in the application.
	Append(ctx context.Context, event *audit.AuditEvent) error
}
