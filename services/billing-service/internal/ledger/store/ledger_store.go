//services/billing-service/internal/ledger/store/ledger_store.go

package ledger_store

import "context"

// LedgerStore defines the interface for the ledger storage layer.
// It specifies methods for recording and retrieving ledger entries.

type LedgerEntry struct {
	EntryID         string // Unique identifier for the ledger entry for idempotency
	TenantID        string // Tenant or Account ID associated with the ledger entry
	TransactionType string // Type of transaction (e.g., "charge", "DEBIT", "CREDIT")
	AmountCents     int64  // Amount in cents (positive for credits, negative for debits)
	Currency        string // Currency code (e.g., "USD")
	Timestamp       int64  // Unix timestamp when the transaction occurred
	Description     string // Description or memo for the ledger entry

}

type LedgerStore interface {
	//CreateLedgerEntry inserts a new ledger row
	// IT is idempotent . IF the  EntryID exists It returns nil ( success ) . But does not create duplicate
	CreateLedgerEntry(ctx context.Context, entry LedgerEntry) error
}
