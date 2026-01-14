//services/billing-service/internal/payment/payment_store.go

package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PaymentAttemptStore handles the persistence of payment attempts.
type PaymentAttemptStore interface {
	//CreateAttempt persists a new payment attempt record in the store.
	// CreateAttempt records the INTENT to pay. Must be called BEFORE Stripe
	CreatePaymentAttempt(ctx context.Context, attempt *PaymentAttempt) error
	// UpdateStatus transitions the state (e.g., PENDING -> SUCCEEDED).
	// It should also update 'provider_payment_id' if it wasn't available at creation.
	UpdateAttemptStatus(ctx context.Context, attemptID uuid.UUID, status PaymentStatus, providerID string, errCode *string, errMsg *string) error
	// GetPendingAttempts fetches "stuck" payments for the Reconciliation Worker.
	// limit: Batch size processing (e.g., process 50 at a time).
	// olderThan: Definition of "stuck" (e.g., created > 5 mins ago)
	GetPendingAttempts(ctx context.Context, limit int, olderThan time.Duration) ([]*PaymentAttempt, error)
	// GetAttemptByProviderID is crucial for Webhooks (Lookup by 'pi_...').
	// It allows us to correlate incoming webhook events with our internal records.it fetches payment attempt using provider payment id
	GetAttemptByProviderID(ctx context.Context, providerID string) (*PaymentAttempt, error)
}
