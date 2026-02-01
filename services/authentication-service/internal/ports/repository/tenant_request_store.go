//services/authentication-service/internal/ports/repository/tenant_request_store.go

package repository

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/google/uuid"
)

type TenantRequestStore interface {
	CreateTntRequest(ctx context.Context, req *tenant.TenantCreationRequest) error
	GetTntRequestByID(ctx context.Context, id uuid.UUID) (*tenant.TenantCreationRequest, error)
	// GetPendingTntRequestByUser ensures one pending request per user (Spam prevention)
	GetPendingTntRequestByUser(ctx context.Context, userID uuid.UUID) (*tenant.TenantCreationRequest, error)
	// Update atomically updates the status (CAS or transaction)
	UpdateTntRequest(ctx context.Context, req *tenant.TenantCreationRequest) error // For status updates
}
