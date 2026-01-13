// services/billing-service/internal/payment/Payment_Service.go
package payment

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/invoice"
	"github.com/google/uuid"
)

//payment service orcastrates payment processing by interacting with various components like AccountProvider, PaymentGateway, InvoiceReaderUpdater, and LedgerRecorder.

type PaymentService struct {
	invoiceReader   InvoiceReader
	invoiceUpdater  InvoiceUpdater
	accountProvider AccountProvider
	paymentGateway  PaymentGateway
	ledgerRecorder  LedgerRecorder

	//singleFlight ensures multiple concurrent requests for the same invoice. Are deduped into a single payment attempt.
	//singleflight.Group: If 50 requests come in for Invoice #101 at the exact same second,
	// we only execute one payment flow. The other 49 wait and receive the same result.
	// This saves 49 Stripe API calls and DB writes.

	sf                  singleflight.Group
	paymentAttemptStore PaymentAttemptStore // Interface to persist payment attempts
}

func NewPaymentService(
	invoiceReader InvoiceReader,
	invoiceUpdater InvoiceUpdater,
	accountProvider AccountProvider,
	paymentGateway PaymentGateway,
	ledgerRecorder LedgerRecorder,
	paymentAttemptStore PaymentAttemptStore,
) *PaymentService {
	return &PaymentService{
		invoiceReader:   invoiceReader,
		invoiceUpdater:  invoiceUpdater,
		accountProvider: accountProvider,
		paymentGateway:  paymentGateway,
		ledgerRecorder:  ledgerRecorder,
		//sf is zero-value initialized, which is safe to use.
		paymentAttemptStore: paymentAttemptStore,
	}
}

// PayInvoice attempts to collect payment.
// Wraps logic in SingleFlight to prevent "Thundering Herd" on the database/Stripe.
func (ps *PaymentService) PayInvoice(ctx context.Context, invoiceID uuid.UUID) error { // SingleFlight Key: "payment_process_<uuid>"
	// 1. Create a unique key for this operation
	key := fmt.Sprintf("pay_invoice_%s", invoiceID.String())

	// 2. Execute via SingleFlight
	// We ignore the return value (val) because we only care if it succeeded or failed.
	_, err, _ := ps.sf.Do(key, func() (interface{}, error) {
		return nil, ps.PayInvoiceExecution(ctx, invoiceID)
	})

	return err

}

// processPaymentLogic contains the actual business logic.
// This is extracted to keep the SingleFlight closure clean.
func (ps *PaymentService) PayInvoiceExecution(ctx context.Context, invoiceID uuid.UUID) error {
	//context Timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Fetch the invoice details

	inv, err := ps.invoiceReader.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return errors.New("failed to fetch invoice: " + err.Error())
	}
	// State Validation (The GateKeeper )
	// We do this check in memory to fail fast, but the DB will do the final check.
	if inv.Status == invoice.InvoicePaid {
		return nil // Idempotency: It's already done. Treat as success.
	}
	if inv.Status != invoice.InvoiceFinalized {
		return fmt.Errorf("invoice is not in FINALIZED state (current: %s)", inv.Status)
	}
	//  Zero-Amount Handling (Edge Case)
	// ⚠️ STRIPE RULE: We CANNOT charge 0 cents. Minimum is usually 50 cents.
	// If the user owes nothing, we skip the gateway and mark it paid internally.
	if inv.TotalCents <= 0 {
		log.Printf("Invoice %s has 0 amount. Skipping Stripe.", inv.InvoiceID)
		return ps.MarkAsPaidInternal(ctx, inv.InvoiceID, inv.TenantID, 0, inv.Currency, "system-zero-amount")
	}
	//Fetch billing account details for the tenant
	account, err := ps.accountProvider.GetBillingAccountDetails(ctx, inv.TenantID)
	if err != nil {
		return errors.New("failed to fetch billing account details: " + err.Error())
	}
	// ---  Record Intent (The State Machine) ---
	attemptID := uuid.New()
	attempt := &PaymentAttempt{
		AttemptID:   attemptID,
		InvoiceID:   inv.InvoiceID,
		TenantID:    inv.TenantID,
		Provider:    "Stripe", // Assuming Stripe for now
		Status:      PaymentStatusPending,
		AmountCents: inv.TotalCents,
		Currency:    inv.Currency,
		//ProviderPaymentID will be set later when we get it from Stripe
	}
	// Persist INTENT before network call
	if err := ps.paymentAttemptStore.CreatePaymentAttempt(ctx, attempt); err != nil {
		return fmt.Errorf("failed to record payment attempt: %w", err)
	}

	//Construct PAyment Request
	req := PaymentRequest{
		ReferenceID:     attemptID.String(), //Use ATTEMPT ID as idempotency key, not Invoice ID. This allows retries!
		AmountCents:     inv.TotalCents,
		Currency:        inv.Currency,
		CustomerID:      account.StripeCustomerID, //Assuming we are using Stripe
		PaymentMethodID: account.PaymentMethodID,  //Assuming we are using Stripe
		Description:     fmt.Sprintf("Invoice #%s for %s", inv.InvoiceID.String(), account.Email),
		MetaData: map[string]string{
			"tenant_id":  inv.TenantID.String(),
			"invoice_id": inv.InvoiceID.String(),
		},
	}
	// Execute Charge with Timeout (Safety Guard)
	// We create a new context with a hard 30s limit for Stripe.
	// This ensures we don't hang forever if Stripe is slow.
	stripeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// Execute Charge (External Phase)
	// This is the slowest part. If context cancels here, Stripe might still charge,
	// but our Idempotency Key protects us on retry.
	result, err := ps.paymentGateway.ChargeAttempt(stripeCtx, req)
	if err != nil {
		// Record FAILURE in DB
		failMsg := err.Error()
		// We ignore the error from UpdateAttemptStatus because the original 'err' is more important to return
		_ = ps.paymentAttemptStore.UpdateAttemptStatus(ctx, attemptID, PaymentFailed, "", nil, &failMsg)
		return fmt.Errorf("payment gateway declined: %w", err)
	}
	// Handle Success ---
	// Update State Machine to SUCCEEDED
	if dbErr := ps.paymentAttemptStore.UpdateAttemptStatus(ctx, attemptID, PaymentSucceeded, result.TransactionID, nil, nil); dbErr != nil {
		// 🚨 CRITICAL: Stripe charged, but we couldn't update the attempt to SUCCEEDED.
		// This leaves the attempt as PENDING. The Reconciler will find it and fix it.
		log.Printf("[CRITICAL] Payment succeeded (Tx: %s) but Attempt Update failed: %v", result.TransactionID, dbErr)
		// We continue! Try to mark invoice paid anyway.
	}
	//  Finalize Transaction (Write Phase)
	// If we crash here, we have a "Ghost Charge" (Paid in Stripe, Unpaid in DB).
	// Handling this strictly requires a background reconciler (Phase 4).
	// For now, we log heavily if this fails.
	return ps.MarkAsPaidInternal(ctx, inv.InvoiceID, inv.TenantID, inv.TotalCents, inv.Currency, result.TransactionID)
}

// markAsPaidInternal handles the local side-effects of a successful payment.
// It updates the Invoice status AND writes to the Ledger.
func (ps *PaymentService) MarkAsPaidInternal(
	ctx context.Context,
	invID uuid.UUID,
	tenantID uuid.UUID,
	amount int64,
	currency string,
	txID string,
) error {
	// A. Update Invoice Status (Primary Truth)
	if err := ps.invoiceUpdater.MarkInvoicePaid(ctx, invID, txID); err != nil {
		// Log critical error: Money moved, but DB didn't update.
		log.Printf("[CRITICAL] Payment %s succeeded but Invoice %s update failed: %v", txID, invID, err)
		return fmt.Errorf("critical: payment succeeded but db update failed: %w", err)
	}

	// B. Record Ledger Entry (Secondary)
	// We use the interface method RecordTransaction (as defined in Step 5)
	description := fmt.Sprintf("Payment for Invoice %s (Ref: %s)", invID.String(), txID)

	err := ps.ledgerRecorder.RecordCreditTransaction(ctx, tenantID, amount, currency, invID.String(), description)
	if err != nil {
		// Non-blocking error. We log it and continue.
		// A background reconciler can fix the ledger later.
		log.Printf("[WARN] Invoice %s paid, but Ledger record failed: %v", invID, err)
	}

	return nil
}
