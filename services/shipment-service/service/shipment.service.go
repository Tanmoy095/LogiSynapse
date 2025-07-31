// service/shipment.go
package service

import (
	"context"
	"errors"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/kafka"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/store"
)

// ShipmentService handles business logic for shipments, using a ShipmentStore for data access.
// Analogy: The chef who uses the pantry (store) to prepare dishes (shipments).
type ShipmentService struct {
	store    store.ShipmentStore // Use interface instead of concrete MemoryStore
	producer *kafka.KafkaProducer
}

// NewShipmentService creates a new service with the given store.
// Analogy: Hires a chef and gives them access to the pantry's menu (interface).
func NewShipmentService(store store.ShipmentStore, producer *kafka.KafkaProducer) *ShipmentService {
	return &ShipmentService{
		store:    store,
		producer: producer,
	}
}

// CreateShipment validates and stores a new shipment.
func (s *ShipmentService) CreateShipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error) {
	// Basic validation
	if shipment.Origin == "" || shipment.Destination == "" {
		return models.Shipment{}, errors.New("missing required fields")
	}
	// Store the shipment using the interface
	created, err := s.store.CreateShipment(ctx, shipment)
	if err != nil {
		return models.Shipment{}, err

	}
	// Create a Kafka event payload with the event type and shipment details.
	// Using map[string]interface{} for flexibility in event structure.
	event := map[string]interface{}{
		"event":   "shipment.created", // Identifies the event type.
		"payload": created,            // Includes shipment details (ID, Origin, Destination, etc.).
	}
	// Publish the event to Kafka in a goroutine for fire-and-forget.
	// Uses the shipment ID as the key to ensure ordered processing in partitions.
	go s.producer.Publish(context.Background(), created.ID, event)

	// Return the created shipment for the gRPC response.
	return created, nil
}

// / GetShipments retrieves shipments based on filters and pagination.
// - origin, status, destination: Filter criteria.
// - limit, offset: Pagination parameters.
func (s *ShipmentService) GetShipments(origin, status, destination string, limit, offset int32) ([]models.Shipment, error) {
	return s.store.GetShipments(context.Background(), origin, status, destination, limit, offset)
}
