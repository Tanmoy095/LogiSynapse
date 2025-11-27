// client/shipment.go
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/graphql-gateway/internal/models"
	"github.com/Tanmoy095/LogiSynapse/shared/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Helper function to process gRPC errors
func handleGRPCError(err error, serviceName string) error {
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, so return it as is
		return err
	}

	// Check for specific gRPC status codes
	switch st.Code() {
	case codes.Unavailable:
		return fmt.Errorf("%s service is unavailable", serviceName)
	case codes.NotFound:
		return fmt.Errorf("resource not found in %s service", serviceName)
	// Add other cases as needed
	default:
		// Return a generic error with the original message
		return fmt.Errorf("gRPC error from %s service: %s", serviceName, st.Message())
	}
}

// ShipmentClient connects to the Shipment Service via gRPC.
type ShipmentClient struct {
	client proto.ShipmentServiceClient
	conn   *grpc.ClientConn // Store connection for graceful shutdown
}

// NewShipmentClient initializes a gRPC client to connect to the Shipment Service.
// It uses a timeout and blocks until connected to ensure the service is available.
func NewShipmentClient(addr string) (*ShipmentClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to shipment service: %v", err)
	}
	return &ShipmentClient{
		client: proto.NewShipmentServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close shuts down the gRPC connection gracefully.
// Analogy: Turns off the intercom when the restaurant closes.
func (c *ShipmentClient) Close() error {
	return c.conn.Close()
}

// GetShipments calls the Shipment Service's GetShipments endpoint.
// It converts the response to local models for GraphQL.
func (c *ShipmentClient) GetShipments(ctx context.Context, origin, destination string, status proto.ShipmentStatus, limit, offset int32) ([]models.Shipment, error) {
	req := &proto.GetShipmentsRequest{
		Origin:      origin,
		Status:      status,
		Destination: destination,
		Limit:       limit,
		Offset:      offset,
	}
	resp, err := c.client.GetShipments(ctx, req)
	if err != nil {
		return nil, handleGRPCError(err, "shipment")

	}

	// Convert proto.Shipment to local models.Shipment..logic same as before...
	ModelShipments := make([]models.Shipment, len(resp.Shipments))
	for i, shipment := range resp.Shipments {

		ModelShipments[i] = models.Shipment{

			ID:          shipment.Id,
			Origin:      shipment.Origin,
			Destination: shipment.Destination,
			Eta:         shipment.Eta,
			Status:      shipment.Status,
			Carrier: models.Carrier{
				Name:        shipment.Carrier.Name,
				TrackingURL: shipment.Carrier.TrackingUrl,
			},
		}

	}
	return ModelShipments, err

}

// CreateShipment calls the Shipment Service's CreateShipment endpoint.
// It converts the input to gRPC format and the response to local models.
func (c *ShipmentClient) CreateShipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error) {

	req := &proto.CreateShipmentRequest{
		Origin:      shipment.Origin,
		Destination: shipment.Destination,
		Eta:         shipment.Eta,
		Status:      shipment.Status, // Convert models.ShipmentStatus to string for gRPC
		Carrier: &proto.Carrier{
			Name:        shipment.Carrier.Name,
			TrackingUrl: shipment.Carrier.TrackingURL,
		},
	}
	resp, err := c.client.CreateShipment(ctx, req)
	if err != nil {
		// Convert gRPC error to status and check code
		s, ok := status.FromError(err)
		if ok && s.Code() == codes.Unavailable {
			return models.Shipment{}, fmt.Errorf("shipment service is unavailable")
		}
		return models.Shipment{}, fmt.Errorf("failed to create shipment: %v", err)
	}
	// Convert gRPC status (string) to models.ShipmentStatus

	return models.Shipment{
		ID:          resp.Shipment.Id,
		Origin:      resp.Shipment.Origin,
		Destination: resp.Shipment.Destination,
		Eta:         resp.Shipment.Eta,
		Status:      resp.Shipment.Status,
		Carrier: models.Carrier{
			Name:        resp.Shipment.Carrier.Name,
			TrackingURL: resp.Shipment.Carrier.TrackingUrl,
		},
	}, nil
}
