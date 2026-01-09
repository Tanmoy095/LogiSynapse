// services/billing-service/internal/accounts/account_store.go
package accounts

import (
	"context"

	"github.com/google/uuid"
)

// AccountStore handles persistence for tenant billing accounts.
type AccountStore interface {
	GetBillingAccountDetails(ctx context.Context, tenantID uuid.UUID) (*Account, error)
}
