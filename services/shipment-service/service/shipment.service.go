package service

import (
	"errors"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/store"
)

//Implement the Business Logic

type ShipmentService struct {
	store *store.MemoryStore
}

func NewShipmenmtService(store *store.MemoryStore) *ShipmentService {
	return &ShipmentService{
		store: store,
	}
}
func (s *ShipmentService) CreateShipment(shipment models.Shipment) (models.Shipment, error) {
	// Basic validation (expand later)
	if shipment.ID == "" || shipment.Origin == "" || shipment.Destination == "" {
		return models.Shipment{}, errors.New("missing required fields")
	}
	err := s.store.Create(shipment)
	if err != nil {
		return models.Shipment{}, err

	}
	return shipment, nil
}
func (s *ShipmentService) GetShipments(origin, status, destination string, limit, offset int32) ([]models.Shipment, error) {
	return s.store.GetAll(origin, status, destination, limit, offset)
}
