//services/billing-service/internal/payment/webhookLogic.payment.go

package payment

import (
	"context"
	"fmt"
	"log"
)

// HandleAsyncResult processes a normalized event from ANY provider.
// It is the bridge between the outside world (Webhook) and our Database.
func (ps *PaymentService) HandleAsyncResult(ctx context.Context, event NormalizedEvent) error {
	log.Printf("[Webhook] Processing %s event for ID: %s", event.Provider, event.ProviderPaymentID)

	// 1. Find the Local Payment Attempt
	// We use the Provider ID (e.g., 'pi_123') to look up our DB record.
	attempt, err := ps.paymentAttemptStore.GetAttemptByProviderID(ctx, event.ProviderPaymentID)
	if err != nil {
		// If we can't find it, it might be an old payment or from another system.
		return fmt.Errorf("attempt not found for provider_id %s: %w", event.ProviderPaymentID, err)
	}

	// 2. Idempotency Check
	// If the DB is already in the final state, we stop.
	if attempt.Status == event.Status {
		return nil
	}
	if attempt.Status == PaymentSucceeded {
		return nil // Never overwrite a Success with a Failure
	}

	// 3. Update the State Machine (payment_attempts table)
	if err := ps.paymentAttemptStore.UpdateAttemptStatus(
		ctx,
		attempt.AttemptID,
		event.Status,
		event.ProviderPaymentID,
		event.ErrorCode,
		event.ErrorMessage,
	); err != nil {
		return fmt.Errorf("failed to update attempt status: %w", err)
	}

	// 4. If Successful, Finalize the Business Transaction
	if event.Status == PaymentSucceeded {
		// This marks the Invoice as PAID and updates the Ledger
		// Reuse the logic we built in Step 2!
		err = ps.FinalizeSuccessfulPayment(
			ctx,
			attempt.InvoiceID,
			attempt.TenantID,
			attempt.AmountCents,
			attempt.Currency,
			event.ProviderPaymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to finalize invoice logic: %w", err)
		}
	}

	return nil
}
