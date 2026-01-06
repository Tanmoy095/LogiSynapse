// services/billing-service/internal/invoice/invoice_generator.go

package invoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/store"
	"github.com/google/uuid"
)

var (
	ErrInvoiceAlreadyFinalized = errors.New("cannot regenerate a finalized invoice")
)

type InvoiceGenerator struct {
	// Dependencies would go here (e.g., stores, calculators)
	//for invoice generation we need ledger store and usage store
	LedgerStore  store.LedgerStore
	InvoiceStore InvoiceStore
}

func NewInvoiceGenerator(
	ledgerStore store.LedgerStore,
	invoiceStore InvoiceStore,
) *InvoiceGenerator {
	return &InvoiceGenerator{
		LedgerStore:  ledgerStore,
		InvoiceStore: invoiceStore,
	}
}

// GenerateInvoiceForTenant aggregates ledger entries into a formal invoice ...

func (ig *InvoiceGenerator) GenerateInvoiceForTenant(ctx context.Context, tenantID uuid.UUID, year int, month int) (*Invoice, error) {

	// 1. Check for existing invoice
	existingInvoice, err := ig.InvoiceStore.GetInvoice(ctx, tenantID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invoice: %w", err)
	}

	if existingInvoice != nil {
		// Rule: Immutable History
		if existingInvoice.Status != InvoiceDraft {
			return nil, ErrInvoiceAlreadyFinalized
		}

		// Rule: If Draft exists, delete it completely and regenerate.
		// This is cleaner than trying to "update" lines.
		if err := ig.InvoiceStore.DeleteInvoice(ctx, existingInvoice.InvoiceID); err != nil {
			return nil, fmt.Errorf("failed to delete old draft: %w", err)
		}
	}
	// 2. Fetch Source of Truth (The Ledger)
	entries, err := ig.LedgerStore.GetEntriesForPeriod(ctx, tenantID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ledger entries: %w", err)
	}
	//If no activity we might still want to generate a $0 invoice or return nil
	//for now lets return nil (no invoice needed )
	if len(entries) == 0 {
		return nil, nil
	}

	// 3. LOGIC: Group Ledger Entries by Usage Type
	// We use a map where Key = "SHIPMENT_CREATED" and Value = Pointer to Line
	lineMap := make(map[billingtypes.UsageType]*InvoiceLine)
	var invoiceTotal int64
	for _, entry := range entries {
		// DEBITS (Charges) for the invoice total and Credit subtraction We owe them (Negative)
		// Calculate the net impact of this entry
		var amount int64
		if entry.TransactionType == billingtypes.TransactionTypeDebit {
			amount = entry.AmountCents // Debits add to invoice total .// We charge them (Positive)
		} else if entry.TransactionType == billingtypes.TransactionTypeCredit {
			amount = -entry.AmountCents // Credits reduce the invoice total ..// We owe them (Negative)
		} else {
			continue // Unknown type, skip
		}

		// update global Invoice Total
		invoiceTotal += amount // Aggregate to invoice total
		// GROUPING STRATEGY:
		// We use the UsageType as the grouping key.
		// If we already have a line for e.g"Shipment Created", we just add to its total.
		// If not, we create a new line.
		if line, exists := lineMap[entry.UsageType]; exists {
			// line exists, add amount just update amount
			line.LineTotalCents += amount
		} else {
			// New line : create and add to map
			lineMap[entry.UsageType] = &InvoiceLine{
				ID:             uuid.New(),
				UsageType:      entry.UsageType,
				Description:    fmt.Sprintf("%s Charges", entry.UsageType), // Generic description
				LineTotalCents: amount,
				Quantity:       0, // Ledger doesn't strictly store Qty, only Money. So we leave it 0 or could infer if needed.

			}
		}

	}
	// 4. Flatten the Map into a Slice for the Invoice Lines . it will be easier to store and process
	finalLines := make([]InvoiceLine, 0, len(lineMap))
	for _, line := range lineMap {
		// Only add lines that aren't zero (unless you want to show $0.00 items)
		if line.LineTotalCents != 0 {
			finalLines = append(finalLines, *line)
		}
	}
	// 5. Create the Invoice Object
	invoice := &Invoice{
		InvoiceID:  uuid.New(),
		TenantID:   tenantID,
		Year:       year,
		Month:      month,
		TotalCents: invoiceTotal,
		Currency:   "USD", // Ideally, currency should be consistent per tenant or derived from ledger entries
		Lines:      finalLines,
		Status:     InvoiceDraft, // Start as Draft
		CreatedAt:  time.Now(),
	}
	// 6. Persist Atomically (Header + Lines)
	if err := ig.InvoiceStore.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	return invoice, nil
}
