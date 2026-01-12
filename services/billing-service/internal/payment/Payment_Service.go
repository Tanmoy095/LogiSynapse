package payment

import (
	"context"
	"errors"
	"fmt"
	"log"

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
}

func NewPaymentService(
	invoiceReader InvoiceReader,
	invoiceUpdater InvoiceUpdater,
	accountProvider AccountProvider,
	paymentGateway PaymentGateway,
	ledgerRecorder LedgerRecorder,
) *PaymentService {
	return &PaymentService{
		invoiceReader:   invoiceReader,
		invoiceUpdater:  invoiceUpdater,
		accountProvider: accountProvider,
		paymentGateway:  paymentGateway,
		ledgerRecorder:  ledgerRecorder,
	}
}

// PayInvoice attempts to collect payment for a finalized invoice.

// It relies on the Database (Optimistic Locking) and Stripe (Idempotency Keys)
// to prevent race conditions (double payments).
func (ps *PaymentService) PayInvoice(ctx context.Context, invoiceID uuid.UUID) error {
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
	if inv.TotalCents <= 0 {
		// Auto-resolve zero-dollar invoices (or negative credit notes in future)
		// For now, we just mark them paid without calling Stripe.
		return ps.MarkAsPaidInternal(ctx, inv.InvoiceID, inv.TenantID, 0, inv.Currency, "system-zero-amount")
	}
	//Fetch billing account details for the tenant
	account, err := ps.accountProvider.GetBillingAccountDetails(ctx, inv.TenantID)
	if err != nil {
		return errors.New("failed to fetch billing account details: " + err.Error())
	}
	//Construct PAyment Request
	req := PaymentRequest{
		ReferenceID:     inv.InvoiceID.String(), // CRITICAL: This is our Idempotency Key
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
	// 5. Execute Charge (External Phase)
	// This is the slowest part. If context cancels here, Stripe might still charge,
	// but our Idempotency Key protects us on retry.
	result, err := ps.paymentGateway.ChargeAttempt(ctx, req)
	if err != nil {
		return errors.New("payment charge attempt failed: " + err.Error())
	}
	if err != nil {
		// If Stripe fails, we return the error. The invoice remains FINALIZED.
		// The user can correct their card and retry.
		return fmt.Errorf("payment gateway declined: %w", err)
	}
	// 6. Finalize Transaction (Write Phase)
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

	// A. Update Invoice Status (Optimistic Locking)
	if err := ps.invoiceUpdater.MarkInvoicePaid(ctx, invID, txID); err != nil {
		// 🚨 CRITICAL ERROR: Money taken, but DB failed.
		log.Printf("[CRITICAL] Payment succeeded (Tx: %s) but DB update failed for Invoice %s: %v", txID, invID, err)
		return fmt.Errorf("critical system error: payment succeeded but status update failed: %w", err)
	}

	// B. Record in Ledger (Double Entry Accounting)
	// We credit the user's balance.
	// If this fails, the invoice is PAID but the ledger is out of sync.
	// This is less critical than the invoice status, but still bad.
	description := fmt.Sprintf("Payment for Invoice %s (Ref: %s)", invID.String(), txID)
	if err := ps.ledgerRecorder.RecordCreditTransaction(ctx, tenantID, amount, currency, invID.String(), description); err != nil {
		log.Printf("[ERROR] Invoice %s marked PAID, but Ledger update failed: %v", invID, err)
		// We do NOT return an error here, because the invoice IS legally paid.
		// We swallow the error and rely on logs/metrics to fix the ledger later.
	}

	return nil
}
