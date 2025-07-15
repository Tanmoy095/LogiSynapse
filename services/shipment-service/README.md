LogiSynapse Shipment Service
Overview
The LogiSynapse Shipment Service is a Go-based microservice designed to manage shipment-related operations in a logistics system. It serves as the backend "kitchen" in the Truss-style architecture, handling business logic and data storage while communicating with a GraphQL gateway via gRPC. This service encapsulates the core functionality for creating and retrieving shipments, providing a scalable and maintainable solution for logistics management.
The service is built with a clean architecture, separating concerns into distinct layers: gRPC handlers, business logic, data storage, and shared models. It currently uses an in-memory store for simplicity, with plans to integrate a persistent database (e.g., Postgres) in future iterations.
Features

gRPC API: Exposes GetShipments and CreateShipment endpoints for efficient internal communication.
Business Logic: Handles shipment creation and retrieval with filtering and pagination support.
In-Memory Storage: Uses a thread-safe in-memory store for managing shipments (to be replaced with Postgres).
Clean Architecture: Separates concerns into handlers, services, stores, and models for maintainability and testability.
Extensibility: Designed to support future enhancements like enum-based status, logging, and database integration.

Project Structure
shipment-service/
├── cmd/
│ └── main.go # Entry point for the gRPC server
├── proto/
│ └── shipment.proto # gRPC service definition
│ └── shipment.pb.go # Generated gRPC code
│ └── shipment_grpc.pb.go # Generated gRPC service code
├── handler/grpc/
│ └── shipment.go # gRPC handler implementation
├── service/
│ └── shipment.go # Business logic for shipments
├── store/
│ └── memory.go # In-memory storage implementation
├── internal/
│ └── models/shipment.go # Shared data models
├── config/
│ └── config.go # Configuration management (planned)
└── Makefile # Build and codegen scripts

Key Components

cmd/main.go: Initializes and runs the gRPC server, binding handlers to the service.
proto/shipment.proto: Defines the gRPC service contract, including GetShipments and CreateShipment methods.
handler/grpc/shipment.go: Implements gRPC endpoints, converting between proto and internal models.
service/shipment.go: Contains business logic for shipment operations, delegating data access to the store.
store/memory.go: Manages shipment data in memory with thread-safe operations.
internal/models/shipment.go: Defines the Shipment and Carrier structs for internal use.

Prerequisites

Go: Version 1.20 or higher
protoc: Protocol Buffers compiler
Go gRPC Tools:
google.golang.org/protobuf/cmd/protoc-gen-go
google.golang.org/grpc/cmd/protoc-gen-go-grpc

Install dependencies:
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

Setup and Installation

Clone the Repository:
git clone https://github.com/Tanmoy095/LogiSynapse.git
cd LogiSynapse/shipment-service

Install Go Modules:
go mod tidy

Generate gRPC Code:
protoc --go_out=. --go-grpc_out=. proto/shipment.proto

Build and Run:
go run cmd/main.go

The gRPC server will start on localhost:50051.

Usage
gRPC Endpoints
The service exposes two gRPC methods defined in proto/shipment.proto:

GetShipments:

Request: GetShipmentsRequest (filters by origin, status, destination; supports limit and offset for pagination)
Response: GetShipmentsResponse (list of shipments)
Example:grpcurl -plaintext -d '{"origin":"NY","limit":10,"offset":0}' localhost:50051 shipment.ShipmentService/GetShipments

CreateShipment:

Request: CreateShipmentRequest (includes origin, destination, eta, status, and carrier)
Response: CreateShipmentResponse (created shipment)
Example:grpcurl -plaintext -d '{"origin":"NY","destination":"LA","eta":"2025-07-20","status":"pending","carrier":{"name":"FedEx","tracking_url":"https://fedex.com"}}' localhost:50051 shipment.ShipmentService/CreateShipment

Testing
Test the gRPC service using grpcurl or integrate it with the GraphQL gateway (see Step 4 of the project roadmap).
