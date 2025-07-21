// client/shipment.go
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Tanmoy095/LogiSynapse/graphql-gateway/internal/models"
	"github.com/Tanmoy095/LogiSynapse/graphql-gateway/proto" // Local proto
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
// Analogy: The waiter's intercom to talk to the kitchen.
type ShipmentClient struct {
	client proto.ShipmentServiceClient
	conn   *grpc.ClientConn // Store connection for graceful shutdown
}

// NewShipmentClient initializes a gRPC client to connect to the Shipment Service.
// It uses a timeout and blocks until connected to ensure the service is available.
// Analogy: Sets up the waiter's intercom, ensuring it connects to the kitchen.
func NewShipmentClient(addr string) (*ShipmentClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),             // Use insecure for local development (use TLS in production)
		grpc.WithBlock(),                // Wait until connected
		grpc.WithTimeout(5*time.Second), // Timeout if connection fails
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
// Analogy: Waiter asks the kitchen for a list of dishes (shipments) matching filters.
func (c *ShipmentClient) GetShipments(ctx context.Context, origin, status, destination string, limit, offset int32) ([]models.Shipment, error) {
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
		var ShipmentStatus models.ShipmentStatus
		switch shipment.Status {
		case "IN_TRANSIT":
			ShipmentStatus = models.ShipmentStatusInTransit

		case "DELIVERED":
			ShipmentStatus = models.ShipmentStatusDelivered
		case "PENDING":
			ShipmentStatus = models.ShipmentStatusPending
		default:
			return nil, fmt.Errorf("invalid shipment status: %s", shipment.Status)
		}

		ModelShipments[i] = models.Shipment{

			ID:          shipment.Id,
			Origin:      shipment.Origin,
			Destination: shipment.Destination,
			Eta:         shipment.Eta,
			Status:      ShipmentStatus,
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
// Analogy: Waiter sends a new dish order to the kitchen and gets the prepared dish.
func (c *ShipmentClient) CreateShipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error) {

	req := &proto.CreateShipmentRequest{
		Origin:      shipment.Origin,
		Destination: shipment.Destination,
		Eta:         shipment.Eta,
		Status:      string(shipment.Status), // Convert models.ShipmentStatus to string for gRPC
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
	var shipmentStatus models.ShipmentStatus
	switch resp.Shipment.Status {
	case "IN_TRANSIT":
		shipmentStatus = models.ShipmentStatusInTransit
	case "DELIVERED":
		shipmentStatus = models.ShipmentStatusDelivered
	case "PENDING":
		shipmentStatus = models.ShipmentStatusPending
	default:
		return models.Shipment{}, fmt.Errorf("invalid shipment status: %s", resp.Shipment.Status)
	}
	return models.Shipment{
		ID:          resp.Shipment.Id,
		Origin:      resp.Shipment.Origin,
		Destination: resp.Shipment.Destination,
		Eta:         resp.Shipment.Eta,
		Status:      shipmentStatus,
		Carrier: models.Carrier{
			Name:        resp.Shipment.Carrier.Name,
			TrackingURL: resp.Shipment.Carrier.TrackingUrl,
		},
	}, nil
}
