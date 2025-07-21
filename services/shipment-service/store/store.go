// store/store.go
package store

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
)

// ShipmentStore defines the interface for the storage layer.
// It specifies methods for retrieving and creating shipments.
// Analogy: The pantry's menu, listing what operations (get, create) the chef can request.
type ShipmentStore interface {
	// GetShipments retrieves shipments filtered by origin (and other fields if needed).
	// ctx allows cancellation and timeouts for database operations.
	GetShipments(ctx context.Context, origin, status, destination string, limit, offset int32) ([]models.Shipment, error)

	// CreateShipment adds a new shipment to the store.
	CreateShipment(ctx context.Context, shipment models.Shipment) error
}
