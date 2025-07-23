// cmd/main.go
package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/config"
	grpcServer "github.com/Tanmoy095/LogiSynapse/shipment-service/handler/grpc"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/proto"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/service"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/store"
	"google.golang.org/grpc"
)

// main is the entry point of the Shipment Service, setting up the database connection, store, service, and gRPC server
func main() {
	// Load configuration from environment variables using the config package
	// This retrieves database credentials (e.g., DB_USER, DB_PASSWORD, DB_HOST)
	cfg := config.LoadConfig()

	// Open a connection to the PostgreSQL database using the connection string from cfg
	// Note: This connection is redundant since NewPostgresStore also opens a connection
	db, err := sql.Open("postgres", cfg.GetDBURL())
	if err != nil {
		// Log and exit if the database connection fails
		log.Fatalf("failed to connect to db: %v", err)
	}
	// Ensure the database connection is closed when the program exits
	defer db.Close()

	// Create a new PostgresStore instance to handle database operations
	// Uses the same connection string to establish a connection
	store, err := store.NewPostgresStore(cfg.GetDBURL())
	if err != nil {
		// Log and exit if the store creation fails (e.g., due to connection issues)
		log.Fatalf("failed to create store: %v", err)
	}
	// Ensure the store's database connection is closed when the program exits
	defer store.Close()

	// Initialize the ShipmentService, which contains the business logic
	// It uses the PostgresStore to interact with the database
	svc := service.NewShipmentService(store)

	// Create a TCP listener on port 50051 for the gRPC server
	// This is where the service will listen for incoming gRPC requests
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		// Log and exit if the listener cannot be created
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register the ShipmentService with the gRPC server
	// The grpcServer.NewShipmentServer wraps the ShipmentService to handle gRPC requests
	proto.RegisterShipmentServiceServer(s, grpcServer.NewShipmentServer(svc))

	// Log that the gRPC server is starting
	log.Println("gRPC server running on :50051")

	// Start the gRPC server and serve requests on the listener
	// This blocks until the server stops or an error occurs
	if err := s.Serve(lis); err != nil {
		// Log and exit if the server fails to start
		log.Fatalf("failed to serve: %v", err)
	}
}
