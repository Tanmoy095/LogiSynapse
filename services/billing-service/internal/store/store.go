// services/billing-service/internal/store/store.go
package store

import (
	"context"

	"github.com/google/uuid"
)

type UsageType string

const (
	ShipmentCreated UsageType = "SHIPMENT_CREATED"
	ApiRequest      UsageType = "API_REQUEST"
)

// UsageRecord represents aggregated usage for a billing period
type UsageRecord struct {
	TenantID uuid.UUID
	Type     UsageType
	Quantity int64
	Year     int
	Month    int
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
