// cmd/main.go
package main

import (
	"log"
	"net"

	grpcServer "github.com/Tanmoy095/LogiSynapse/shipment-service/handler/grpc"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/proto"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/service"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/store"
	"google.golang.org/grpc"
)

func main() {
	// Initialize the in-memory store (implements ShipmentStore interface)
	// Analogy: Set up the pantry shelf for the chef to use
	store := store.NewMemoryStore()

	// Initialize the service with the store
	svc := service.NewShipmentService(store)

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterShipmentServiceServer(s, grpcServer.NewShipmentServer(svc))

	log.Println("gRPC server running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
