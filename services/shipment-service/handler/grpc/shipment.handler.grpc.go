package grpcServer

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/proto"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/service"
)

/* gRPC server struct that implements the ShipmentService interface.....
type ShipmentServiceClient interface {
	GetShipments(ctx context.Context, in *GetShipmentsRequest, opts ...grpc.CallOption) (*GetShipmentsResponse, error)
	CreateShipment(ctx context.Context, in *CreateShipmentRequest, opts ...grpc.CallOption) (*CreateShipmentResponse, error)
}
*/

// ShipmentServer is the gRPC server struct that implements the ShipmentService interface
// defined in proto/shipment_grpc.pb.go. It holds a reference to the business logic (service)
type ShipmentServer struct {
	proto.UnimplementedShipmentServiceServer                          // Embeds the default implementation to satisfy the interface
	service                                  *service.ShipmentService // Reference to the business logic layer
}

// NewShipmentServer creates a new ShipmentServer instance, injecting the business logic service.
// Analogy: Sets up a chef (handler) in the kitchen, giving them access to the recipe book (service).
func NewShipmentServer(svc *service.ShipmentService) *ShipmentServer {
	return &ShipmentServer{service: svc}
}

// GetShipments handles the gRPC GetShipments request.
// It receives a request from the waiter (GraphQL gateway via gRPC), fetches shipments
// from the business logic, converts them to gRPC format, and returns the response.

func (s *ShipmentServer) GetShipments(ctx context.Context, req *proto.GetShipmentsRequest) (*proto.GetShipmentsResponse, error) {
	// Call the business logic (service) to fetch shipments based on filters (origin, status, destination)
	// and pagination (limit, offset). The service returns internal models.Shipment structs.
	shipments, err := s.service.GetShipments(req.Origin, req.Status, req.Destination, req.Limit, req.Offset)
	if err != nil {
		return nil, err

	}
	//Convert internal models.Shipment to proto.Shipment for gRPC response
	// Create a slice to hold the converted shipments
	proitoShipments := make([]*proto.Shipment, len(shipments))
	for i, shipment := range shipments {
		// Convert each internal shipment to gRPC-compatible proto.Shipment
		proitoShipments[i] = toProtoShipment(shipment)

	}
	/// Return the gRPC response with the list of converted shipments
	return &proto.GetShipmentsResponse{Shipments: proitoShipments}, nil
}

// CreateShipment handles the gRPC CreateShipment request.
// It receives a new shipment request, converts it to the internal model,
// calls the business logic to create it, and returns the created shipment in gRPC format

func (s *ShipmentServer) CreateShipment(ctx context.Context, req *proto.CreateShipmentRequest) (*proto.CreateShipmentResponse, error) {
	// Convert the gRPC request (proto.CreateShipmentRequest) to an internal models.Shipment
	shipment := toModelShipment(req)
	// Call the business logic to create the shipment (includes validation and storage)
	created, err := s.service.CreateShipment(ctx, shipment)
	if err != nil {
		return nil, err

	}
	// Convert the created shipment back to proto.Shipment for the gRPC response
	return &proto.CreateShipmentResponse{Shipment: toProtoShipment(created)}, nil

}

// toProtoShipment converts an internal models.Shipment to a gRPC proto.Shipment.
// This ensures the response uses the gRPC contract defined in shipment.proto.
func toProtoShipment(s models.Shipment) *proto.Shipment {
	return &proto.Shipment{
		Id:          s.ID,
		Origin:      s.Origin,
		Destination: s.Destination,
		Eta:         s.ETA,
		Status:      s.Status,
		Carrier: &proto.Carrier{
			Name:        s.Carrier.Name,
			TrackingUrl: s.Carrier.TrackingURL,
		},
	}
}

// toModelShipment converts a gRPC proto.CreateShipmentRequest to an internal models.Shipment.
// It generates a unique ID for the new shipment since the request doesn't include one.
func toModelShipment(req *proto.CreateShipmentRequest) models.Shipment {
	return models.Shipment{
		// Generate a unique ID for the shipment
		ID:          "",
		Origin:      req.Origin,
		Destination: req.Destination,
		ETA:         req.Eta,
		Status:      req.Status,
		Carrier: models.Carrier{
			Name:        req.Carrier.Name,
			TrackingURL: req.Carrier.TrackingUrl,
		},
	}
}

// generateID creates a unique ID for a new shipment.
// This is a placeholder; in production, use a robust UUID library (e.g., github.com/google/uuid)
