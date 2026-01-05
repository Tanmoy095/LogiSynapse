package usage

import (
	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/google/uuid"
)

// Shared types
type UsageEvent struct {
	ID        string    //Unique string (e.g., "event-123") for idempotency
	TenantID  uuid.UUID // Tenant or Account ID associated with the usage event
	Type      billingtypes.UsageType
	Quantity  int64 // Quantitative measure of the usage (e.g., number of shipments created)
	Timestamp int64 // Unix timestamp when the
	//  usage occurred
}
