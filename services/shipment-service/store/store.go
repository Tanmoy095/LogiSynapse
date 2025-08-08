// store/store.go
package store

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
)

// ShipmentStore defines the interface for the storage layer.
// It specifies methods for retrieving and creating shipments.
// Specifies Method for crud operations

type ShipmentStore interface {

	// ctx allows cancellation and timeouts for database operations.
	//GetShipments retrieves shipments filtered by origin, status, or destination with pagination.
	GetShipments(ctx context.Context, origin, status, destination string, limit, offset int32) ([]models.Shipment, error)
	//get
	GetShipment(ctx context.Context, id string) (models.Shipment, error)

	// CreateShipment adds a new shipment to the store.
	CreateShipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error)
	UpdateShipment(ctx context.Context, shipment models.Shipment) error
}
