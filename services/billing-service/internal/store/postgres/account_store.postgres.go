// services/billing-service/internal/store/postgres/account_store.postgres.go
package PostgresStore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/accounts"
	"github.com/google/uuid"
)

type AccountStore struct {
	db *sql.DB
}

func NewAccountStore(db *sql.DB) *AccountStore {
	return &AccountStore{db: db}
}

// GetBillingAccountDetails fetches stripe info.
// This method implicitly satisfies 'payment.AccountProvider'.
func (s *AccountStore) GetBillingAccountDetails(ctx context.Context, tenantID uuid.UUID) (*accounts.Account, error) {
	query := `
		SELECT id, email, stripe_customer_id, payment_method_id, current_plan
		FROM accounts
		WHERE id = $1
	`

	var acc accounts.Account
	// Scan directly into struct
	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(
		&acc.ID,
		&acc.Email,
		&acc.StripeCustomerID,
		&acc.PaymentMethodID,
		&acc.CurrentPlan,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account not found for tenant %s", tenantID)
		}
		return nil, fmt.Errorf("db: account fetch failed: %w", err)
	}

	// Data Integrity Check
	if acc.StripeCustomerID == "" || acc.PaymentMethodID == "" {
		return nil, fmt.Errorf("billing details incomplete for tenant %s", tenantID)
	}

	return &acc, nil
}
