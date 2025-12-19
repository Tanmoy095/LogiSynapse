package usage

import "github.com/docker/distribution/uuid"

type UsagesType string

const (
	ShipmentCreated UsagesType = "SHIPMENT_CREATED"
	APIRequest      UsagesType = "API_REQUEST"
)

type UsagesEvent struct {
	id        string    //Unique string (e.g., "event-123") for idempotency
	TenantID  uuid.UUID // Tenant or Account ID associated with the usage event
	Type      UsagesType
	Quantity  int64 // Quantitative measure of the usage (e.g., number of shipments created)
	Timestamp int64 // Unix timestamp when the usage occurred
}
