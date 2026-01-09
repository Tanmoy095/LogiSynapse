//services/billing-service/internal/invoice/invoice_finalizer.go

package invoice

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Finalizer is  responsible for finalizing invoices.
type InvoiceFinalizer struct {
	InvoiceStore InvoiceStore
}

func NewInvoiceFinalizer(store InvoiceStore) *InvoiceFinalizer {
	return &InvoiceFinalizer{
		InvoiceStore: store,
	}
}

// FinalizeInvoice transitions an invoice from DRAFT -> FINALIZED.
// It enforces strict state and integrity checks.
func (inf *InvoiceFinalizer) FinalizeInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	//Fetch Invoice by ID
	// We need to see the current state before we try to change it.
	inv, err := inf.InvoiceStore.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to fetch invoice: %w", err)
	}
	// 2. State Validation (The Gatekeeper)
	// Rule: Only DRAFT invoices can be finalized.
	if inv.Status != InvoiceDraft {
		if inv.Status == InvoiceFinalized {
			return ErrInvoiceAlreadyFinalized
		}
		// Handles PAID, VOID, or unknown states
		return fmt.Errorf("%w: current status is %s", ErrInvoiceNotDraft, inv.Status)
	}
	// 3. Integrity Validation (Sanity Checks)conc
	// Rule: Do not finalize a corrupted or incomplete invoice.
	if inv.TotalCents < 0 {
		return fmt.Errorf("integrity violation: negative total amount %d", inv.TotalCents)
	}
	if inv.Currency == "" {
		return fmt.Errorf("integrity violation: missing currency")
	}

	// 4. Atomic State Transition (Write Phase)
	// We call the store to perform the "Compare-and-Swap" (UPDATE ... WHERE status = DRAFT).
	if err := inf.InvoiceStore.FinalizeInvoice(ctx, invoiceID); err != nil {
		return fmt.Errorf("finalization failed: %w", err)
	}

	// 5. Success
	return nil

}
