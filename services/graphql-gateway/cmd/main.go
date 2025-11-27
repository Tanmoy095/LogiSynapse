// main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Tanmoy095/LogiSynapse/services/graphql-gateway/client"          // gRPC client
	"github.com/Tanmoy095/LogiSynapse/services/graphql-gateway/graph"           // GraphQL resolvers
	"github.com/Tanmoy095/LogiSynapse/services/graphql-gateway/graph/generated" // Generated GraphQL schema
)

// main starts the GraphQL server and connects to the Shipment Service.
// Analogy: Opens the restaurant, sets up the waiter's intercom, and starts serving customers.
func main() {
	// Get Shipment Service address from environment variable (or default)
	addr := os.Getenv("SHIPMENT_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	// Initialize gRPC client to connect to Shipment Service
	// Analogy: Set up the waiter's intercom to call the kitchen
	shipmentClient, err := client.NewShipmentClient(addr)
	if err != nil {
		log.Fatalf("failed to connect to shipment service: %v", err)
	}
	defer shipmentClient.Close() // Close connection when server stops

	// Initialize GraphQL resolver with gRPC client
	resolver := graph.NewResolver(shipmentClient)

	// Set up GraphQL endpoint at /query
	// Analogy: Set up the dining room's service counter for customer orders
	http.Handle("/query", handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver})))

	// Set up GraphiQL playground at root (/) for easy testing
	// Analogy: Provide a menu board for customers to write their orders
	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	log.Println("GraphQL server running on :8080/query")
	log.Println("GraphiQL playground available at :8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
