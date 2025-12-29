//services/billing-service/internal/store/store.go

package store

import (
	"context"
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/google/uuid"
)

// UsageRecord represents aggregated usage for a billing period
type UsageRecord struct {
	TenantID      uuid.UUID
	UsageType     billingtypes.UsageType
	TotalQuantity int64
	BillingPeriod BillingPeriod
}
type BillingPeriod struct {
	Year  int
	Month int
}

// FlushBatch represents a single idempotent flush operation
type FlushBatch struct {
	BatchID uuid.UUID // Idempotency key for the flush
	Records []UsageRecord
}

// UsageStore persists aggregated usage durably
type UsageStore interface {
	// Flush atomically persists a batch.
	// Must be idempotent by BatchID.
	Flush(ctx context.Context, batch FlushBatch) error
	// GetUsageForPeriod  fetches usage records for a tenant for a specific billing period
	GetUsageForPeriod(ctx context.Context, year int, month int) ([]UsageRecord, error)
}

// --- Pricing Store Interface ---

// priceRule represents a pricing rule for different usage tiers
type PriceRule struct {
	TenantID       *uuid.UUID // nil means default rule
	UsageType      billingtypes.UsageType
	UnitPriceCents int64  // price per unit in cents
	Currency       string // currency code, e.g., "USD"
}

type PricingStore interface {
	// GetPriceRules retrieves pricing rules for a given plan
	// GetPrice returns the price rule active at a specific point in time ('at').
	// Strategy: It looks for a tenant-specific price first. If none found, looks for default (tenant_id IS NULL).
	GetPriceRules(ctx context.Context, usageType billingtypes.UsageType, tenantID uuid.UUID, at time.Time) (PriceRule, error)
}

// --- Ledger Store Interface ---

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
