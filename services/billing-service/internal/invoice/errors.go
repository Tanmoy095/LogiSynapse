//services/billing-service/internal/invoice/errors.go

package invoice

import "errors"

var (
	// ErrInvoiceNotFound matches standard 404 behavior
	ErrInvoiceNotFound = errors.New("invoice not found")

	// ErrInvoiceNotDraft protects the state machine.
	// Only DRAFT invoices can be finalized.
	ErrInvoiceNotDraft = errors.New("invoice is not in DRAFT state, cannot finalize")

	// ErrInvoiceAlreadyFinalized is a specific case of NotDraft, helpful for idempotency checks.
	ErrInvoiceAlreadyFinalized = errors.New("invoice is already finalized")

	// ErrEmptyInvoice prevents generating legal docs with no line items.
	ErrEmptyInvoice = errors.New("invoice has no line items")
)
