package store

import (
	"context"
	"sync"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
)

type MemoryStore struct {
	shipments map[string]models.Shipment
	mu        sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		shipments: make(map[string]models.Shipment),
	}
}

func (s *MemoryStore) CreateShipment(ctx context.Context, shipment models.Shipment) error {
	// Check if the context is canceled or timed out
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shipments[shipment.ID] = shipment
	return nil

}
func (s *MemoryStore) GetShipments(ctx context.Context, origin, status, destination string, limit, offset int32) ([]models.Shipment, error) {
	// Check if the context is canceled or timed out
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []models.Shipment
	for _, shipment := range s.shipments {
		if (origin == "" || shipment.Origin == origin) &&
			(status == "" || shipment.Status == status) &&
			(destination == "" || shipment.Destination == destination) {
			result = append(result, shipment)
		}
	}

	// Apply pagination
	start := int(offset)
	end := start + int(limit)
	if start > len(result) {
		return nil, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], nil
}
