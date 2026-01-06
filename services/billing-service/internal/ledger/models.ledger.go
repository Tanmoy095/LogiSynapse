//services/billing-service/internal/ledger/models.ledger.go

package ledger

import (
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/google/uuid"
)

// // TransactionType defines the direction of money movement in the ledger
// type TransactionType string

// // THis enum ensures type safety for transaction types valid only value (credit,debit) can be unused
// // why enum ? I s prevents starting typos like "credit" vs "Credit" which can lead to bugs
// const (
// 	Credit TransactionType = "CREDIT"
// 	Debit  TransactionType = "DEBIT"
// )

type LedgerEntry struct {
	EntryID         string                       // Unique identifier for the ledger entry for idempotency
	TenantID        uuid.UUID                    // Tenant or Account ID associated with the ledger entry
	AmountCents     int64                        // Amount in cents (positive for credits, negative for debits)
	TransactionType billingtypes.TransactionType // Type of transaction (e.g., "charge", "DEBIT", "CREDIT")
	Description     string                       // Description or memo for the ledger entry
	ReferenceID     string                       // External ID (e.g., stripe payment Intent ID)

	Currency  string    // Currency code (e.g., "USD")
	CreatedAt time.Time // Timestamp of when the entry was created
	UsageType billingtypes.UsageType
}
