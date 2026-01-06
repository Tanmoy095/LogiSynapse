package store

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/invoice"
	"github.com/google/uuid"
)

// InvoiceStore defines the interface for the invoice storage layer.
// It specifies methods for creating, retrieving, and updating invoices.
type InvoiceStore interface {
	// CreateInvoice inserts a new invoice into the store.
	CreateInvoice(ctx context.Context, inv invoice.Invoice) error
	// GetInvoice retrieves an invoice by tenant ID, year, and month.
	GetInvoice(ctx context.Context, tenantID uuid.UUID, year int, month int) (invoice.Invoice, error)
	// UpdateStatus updates an existing invoice in the store.
	UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status invoice.InvoiceStatus) error
}
