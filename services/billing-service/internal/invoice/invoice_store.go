// services/billing-service/internal/invoice/invoice_store.go

package invoice

import (
	"context"

	"github.com/google/uuid"
)

// InvoiceStore handles persistence operations for invoices.
// Placed in the invoice package to avoid import cycles between store and invoice.
type InvoiceStore interface {
	CreateInvoice(ctx context.Context, inv *Invoice) error
	GetInvoice(ctx context.Context, tenantID uuid.UUID, year int, month int) (*Invoice, error)
	DeleteInvoice(ctx context.Context, invoiceID uuid.UUID) error
	UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status InvoiceStatus) error
	// GetInvoiceByID fetches a specific invoice by its UUID (not just by tenant/period)
	GetInvoiceByID(ctx context.Context, invoiceID uuid.UUID) (*Invoice, error)
	// FinalizeInvoice performs the state transition from DRAFT -> FINALIZED.
	// It must enforce the condition: WHERE status = 'DRAFT'.
	FinalizeInvoice(ctx context.Context, invoiceID uuid.UUID) error

	//-------------------------------!!After Payment Phase !!-----------------------------------
	// Atomic State Transition for payment marking.
	// Here Atomic state means we change status from FINALIZED to PAID only if current status is FINALIZED

	MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, transactionID string) error
}
