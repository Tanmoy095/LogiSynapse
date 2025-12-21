//services/billing-service/internal/ledger/models.ledger.go

package ledger

import (
	"time"

	"github.com/google/uuid"
)

// TransactionType defines the direction of money movement in the ledger
type TransactionType string

// THis enum ensures type safety for transaction types valid only value (credit,debit) can be unused
// why enum ? I s prevents starting typos like "credit" vs "Credit" which can lead to bugs
const (
	Credit TransactionType = "CREDIT"
	Debit  TransactionType = "DEBIT"
)

type LedgerEntry struct {
	ID          string          // Unique identifier for the ledger entry
	AccountID   string          // Reference to the associated billing account
	AmountCents int64           // always integer .. $10 = 1000 cents
	Type        TransactionType // CREDIT or DEBIT
	Description string          // Description of the transaction mothly subscription or "	Shipment #123"
	ReferenceID string          // External ID (e.g., stripe payment Intent ID)
	CreatedAt   time.Time       // Timestamp of when the entry was created

}

// NewLedgerEntry creates a new ledger entry with the provided details.
func NewLedgerEntry(accountID uuid.UUID, amountCents int64, trnTy TransactionType, desc, refID string) *LedgerEntry {
	return &LedgerEntry{
		ID:          uuid.New().String(),
		AccountID:   accountID.String(),
		AmountCents: amountCents,
		Type:        trnTy,
		Description: desc,
		ReferenceID: refID,
		CreatedAt:   time.Now(),
	}
}
