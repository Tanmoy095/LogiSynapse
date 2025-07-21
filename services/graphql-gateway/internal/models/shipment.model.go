package models

// Carrier represents a shipping carrier
type Carrier struct {
	Name        string
	TrackingURL string
}
type ShipmentStatus string

// Define enum values to match GraphQL schema
const (
	ShipmentStatusInTransit ShipmentStatus = "IN_TRANSIT"
	ShipmentStatusDelivered ShipmentStatus = "DELIVERED"
	ShipmentStatusPending   ShipmentStatus = "PENDING"
)

// Shipment represents a shipment entity
type Shipment struct {
	ID          string
	Origin      string
	Destination string
	Eta         string
	Status      ShipmentStatus
	Carrier     Carrier
}

// CreateShipmentInput defines the input for creating a shipment (for GraphQL)
type CreateShipmentInput struct {
	Origin      string
	Destination string
	Eta         string
	Status      ShipmentStatus
	Carrier     Carrier
}
