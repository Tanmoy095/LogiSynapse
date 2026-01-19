// services/billing-service/internal/store/postgres/payment_attempt_store.go
package PostgresStore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/payment"
	"github.com/google/uuid"
)

type PaymentAttemptStore struct {
	db *sql.DB
}

func NewPaymentAttemptStore(db *sql.DB) *PaymentAttemptStore {
	return &PaymentAttemptStore{db: db}
}

// CreateAttempt persists a new payment attempt record in the store.
func (pa *PaymentAttemptStore) CreatePaymentAttempt(ctx context.Context, attempt *payment.PaymentAttempt) error {

	query := `INSERT INTO payment_attempts 
	(attempt_id, invoice_id, tenant_id, provider, status, amount_cents, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	// Note: provider_payment_id might be empty initially if we haven't called Stripe yet.
	_, err := pa.db.ExecContext(ctx, query,
		attempt.AttemptID,
		attempt.InvoiceID,
		attempt.TenantID,
		attempt.Provider,
		attempt.Status,
		attempt.AmountCents,
		attempt.Currency,
	)
	if err != nil {
		return fmt.Errorf("db: failed to create payment attempt: %w", err)
	}
	return nil

}

// UpdateAttemptStatus transitions the state (e.g., PENDING -> SUCCEEDED).
func (pa *PaymentAttemptStore) UpdateAttemptStatus(ctx context.Context, attemptID uuid.UUID, status payment.PaymentStatus, providerID string, errCode *string, errMsg *string) error {
	query := `
		UPDATE payment_attempts
		SET status = $1, 
		    provider_payment_id = COALESCE(NULLIF($2, ''), provider_payment_id), -- Only update if new value provided
		    error_code = $3,
		    error_message = $4,
		    updated_at = NOW()
		WHERE attempt_id = $5 AND status = 'PENDING'
	`

	_, err := pa.db.ExecContext(ctx, query, status, providerID, errCode, errMsg, attemptID)
	if err != nil {
		return fmt.Errorf("db: failed to update payment attempt status: %w", err)
	}
	return nil

}

// GetPendingAttempts fetches "stuck" payments for the Reconciliation Worker.
// GetPendingAttempts helps in fetching payment attempts that are still in PENDING status
//
//	for longer than the specified duration.
func (pa *PaymentAttemptStore) GetPendingAttempts(ctx context.Context, limit int, olderThan time.Duration) ([]payment.PaymentAttempt, error) {
	// Logic: Find PENDING items created BEFORE (Now - 5 minutes)
	cutOffTime := time.Now().Add(-olderThan) // Calculate cutoff time
	query := `
		SELECT attempt_id, invoice_id, tenant_id, provider, provider_payment_id, status, amount_cents, currency, created_at
		FROM payment_attempts
		WHERE status = 'PENDING' AND created_at < $1 
		ORDER BY created_at ASC -- Process oldest first (FIFO)
		LIMIT $2
	`
	rows, err := pa.db.QueryContext(ctx, query, cutOffTime, limit)
	if err != nil {
		return nil, fmt.Errorf("db: failed to fetch pending payment attempts: %w", err)
	}
	defer rows.Close()

	var attempts []payment.PaymentAttempt
	for rows.Next() {
		var paymentAttempt payment.PaymentAttempt
		var providerPaymentID sql.NullString // Handle nullable field . it helps in scanning nullable string from db
		err := rows.Scan(
			&paymentAttempt.AttemptID,
			&paymentAttempt.InvoiceID,
			&paymentAttempt.TenantID,
			&paymentAttempt.Provider,
			&providerPaymentID,
			&paymentAttempt.Status,
			&paymentAttempt.AmountCents,
			&paymentAttempt.Currency,
			&paymentAttempt.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("db: failed to scan payment attempt: %w", err)
		}
		if providerPaymentID.Valid {
			paymentAttempt.ProviderPaymentID = providerPaymentID.String // it means value is not null, assign it

		}
		attempts = append(attempts, paymentAttempt)

	}

	return attempts, nil
}

// GetAttemptByProviderID is crucial for Webhooks (Lookup by 'pi_...').
// It allows us to correlate incoming webhook events with our internal records.
func (pa *PaymentAttemptStore) GetAttemptByProviderID(ctx context.Context, providerID string) (*payment.PaymentAttempt, error) {
	query := `
		SELECT attempt_id, invoice_id, tenant_id, status, amount_cents
		FROM payment_attempts
		WHERE provider_payment_id = $1
	`
	var payment payment.PaymentAttempt
	err := pa.db.QueryRowContext(ctx, query, providerID).Scan(&payment.AttemptID, &payment.InvoiceID, &payment.TenantID, &payment.Status, &payment.AmountCents)
	if err != nil {
		return nil, err // Let caller handle ErrNoRows
	}
	return &payment, nil

}
