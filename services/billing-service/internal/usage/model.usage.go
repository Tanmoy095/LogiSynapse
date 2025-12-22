//services/billing-service/internal/usage/model.usage.go

package usage

import "github.com/google/uuid"

type UsageType string

const (
	ShipmentCreated UsageType = "SHIPMENT_CREATED"
	APIRequest      UsageType = "API_REQUEST"
)

type UsageEvent struct {
	id        string    //Unique string (e.g., "event-123") for idempotency
	TenantID  uuid.UUID // Tenant or Account ID associated with the usage event
	Type      UsageType
	Quantity  int64 // Quantitative measure of the usage (e.g., number of shipments created)
	Timestamp int64 // Unix timestamp when the usage occurred
}
