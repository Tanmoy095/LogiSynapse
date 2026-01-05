package invoice

import (
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
	InvoiceStore store.InvoiceStore
}

func NewInvoiceGenerator(
	ledgerStore store.LedgerStore,
	invoiceStore store.InvoiceStore,
) *InvoiceGenerator {
	return &InvoiceGenerator{
		LedgerStore:  ledgerStore,
		InvoiceStore: invoiceStore,
	}
}

// GenerateInvoiceForTenant aggregates ledger entries into a formal invoice ...

func (ig *InvoiceGenerator) GenerateInvoiceForTenant(tenantID uuid.UUID, year int, month int) (*Invoice, error) {

	// 1. Check if invoice exists

	existingInvoice, err := ig.InvoiceStore.GetInvoice(tenantID, year, month)
	if err == nil && existingInvoice != nil { // err== nil means invoice exists . invoice!= nil means invoice found
		if existingInvoice.Status != InvoiceDraft { // Not Draft
			// Rule: Immutable History for Finalized/Paid/Voided
			return nil, ErrInvoiceAlreadyFinalized
		}

		// If Draft, we  delete and regenerate, or return existing.
		// we will delete existing draft and regenerate lines and re-run

		err = ig.InvoiceStore.DeleteInvoice(existingInvoice.ID)
		if err != nil {
			return nil, err
		}
		//regenerate below

	}
	// 2. Fetch Source of Truth (The Ledger)
	entries, err := ig.LedgerStore.GetLedgerEntriesForPeriod(tenantID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ledger entries: %w", err)
	}
	if len(entries) == 0 {
		return nil, errors.New("no ledger entries found for period")
	}

	// 3. LOGIC: Group Ledger Entries by Usage Type
	// We use a map where Key = "SHIPMENT_CREATED" and Value = Pointer to Line
	lineMap := make(map[billingtypes.UsageType]*InvoiceLine)
	var invoiceTotal int64
	for _, entry := range entries {
		// DEBITS (Charges) for the invoice total and Credit subtraction We owe them (Negative)
		// Calculate the net impact of this entry
		var amount int64
		if entry.TransactionType == string(billingtypes.TransactionTypeDebit) {
			amount = entry.AmountCents // Debits add to invoice total .// We charge them (Positive)
		} else if entry.TransactionType == string(billingtypes.TransactionTypeCredit) {
			amount = -entry.AmountCents // Credits reduce the invoice total ..// We owe them (Negative)
		} else {
			continue // Unknown type, skip
		}

		// update global Invoice Total
		invoiceTotal += amount // Aggregate to invoice total
		// GROUPING STRATEGY:
		// We use the UsageType as the grouping key.
		// If we already have a line for "Shipment Charges", we just add to its total.
		// If not, we create a new line.
		key := entry.UsageType // Simplified mapping means we assume TransactionType == UsageType
		if line, exists := lineMap[billingtypes.UsageType(key)]; exists {
			// line exists, add amount just update amount
			line.LineTotalCents += amount
		} else {
			// New line : create and add to map
			lineMap[billingtypes.UsageType(key)] = &InvoiceLine{
				ID:             uuid.New(),
				UsageType:      billingtypes.UsageType(key),
				Description:    entry.Description,
				LineTotalCents: amount,
				Quantity:       0, // Quantity is not tracked in ledger entries directly

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
		Currency:   "USD",
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
