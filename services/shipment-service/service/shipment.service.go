// service/shipment.go
package service

import (
	"context"
	"errors"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/store"
)

// ShipmentService handles business logic for shipments, using a ShipmentStore for data access.
// Analogy: The chef who uses the pantry (store) to prepare dishes (shipments).
type ShipmentService struct {
	store store.ShipmentStore // Use interface instead of concrete MemoryStore
}

// NewShipmentService creates a new service with the given store.
// Analogy: Hires a chef and gives them access to the pantry's menu (interface).
func NewShipmentService(store store.ShipmentStore) *ShipmentService {
	return &ShipmentService{store: store}
}

// CreateShipment validates and stores a new shipment.
func (s *ShipmentService) CreateShipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error) {
	// Basic validation
	if shipment.Origin == "" || shipment.Destination == "" {
		return models.Shipment{}, errors.New("missing required fields")
	}
	// Store the shipment using the interface
	return s.store.CreateShipment(ctx, shipment)
}

// GetShipments retrieves shipments based on filters and pagination.
// Analogy: Chef asks the pantry for shipments matching the criteria.
func (s *ShipmentService) GetShipments(origin, status, destination string, limit, offset int32) ([]models.Shipment, error) {
	return s.store.GetShipments(context.Background(), origin, status, destination, limit, offset)
}
