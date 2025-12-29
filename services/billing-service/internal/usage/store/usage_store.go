// services/billing-service/internal/usage/store/usage_store.go
package store

import (
	"context"

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
	BatchID string // Idempotency key for the flush
	Records []UsageRecord
}

// UsageStore persists aggregated usage durably
type UsageStore interface {
	// Flush atomically persists a batch.
	// Must be idempotent by BatchID.
	Flush(ctx context.Context, batch FlushBatch) error
}
