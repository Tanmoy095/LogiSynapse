package kafka

// ShipmentCreatedEvent represents the structure of the Kafka event from shipment-service.
// It matches the map[string]interface{} payload sent by the producer.

type shipmentCreatedEvent struct {
	Event   string                 `json:"event"`   // Event type (e.g., "shipment.created").
	Payload map[string]interface{} `json:"payload"` //Shipment details (ID, Origin, Destination, etc.).
}
