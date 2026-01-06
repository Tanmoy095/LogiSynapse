package store

import (
	"context"
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/ledger"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/pricing"
	"github.com/google/uuid"
)

// UsageStore persists aggregated usage durably
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

type UsageStore interface {
	// Flush atomically persists a batch.
	// Must be idempotent by BatchID.
	Flush(ctx context.Context, batch FlushBatch) error
	// GetUsageForPeriod  fetches usage records for a tenant for a specific billing period
	GetUsageForPeriod(ctx context.Context, year int, month int) ([]UsageRecord, error)
}

// --- Pricing Store Interface ---

type PricingStore interface {
	// GetPriceRules retrieves pricing rules for a given plan
	// GetPrice returns the price rule active at a specific point in time ('at').
	// Strategy: It looks for a tenant-specific price first. If none found, looks for default (tenant_id IS NULL).
	GetPriceRules(ctx context.Context, usageType billingtypes.UsageType, tenantID uuid.UUID, at time.Time) (pricing.PriceRule, error)
}

// --- Ledger Store Interface ---........................................

type LedgerStore interface {
	//CreateLedgerEntry inserts a new ledger row
	// IT is idempotent . IF the  EntryID exists It returns nil ( success ) . But does not create duplicate
	CreateLedgerEntry(ctx context.Context, entry ledger.LedgerEntry) error
	GetEntriesForPeriod(ctx context.Context, tenantID uuid.UUID, year int, month int) ([]ledger.LedgerEntry, error)
}

// InvoiceStore interface has been moved to the invoice package to avoid an import cycle.
