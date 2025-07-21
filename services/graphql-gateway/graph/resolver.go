package graph

// Resolver serves as dependency injection container for your app
// graph/resolver.go

import (
	"github.com/Tanmoy095/LogiSynapse/graphql-gateway/client" // gRPC client
)

// Resolver holds dependencies for GraphQL resolvers.
// Analogy: The waiter, who uses the intercom (gRPC client) to talk to the kitchen.
type Resolver struct {
	shipmentClient *client.ShipmentClient
}

// NewResolver initializes the resolver with a gRPC client.
// Analogy: Hires a waiter and gives them the intercom to contact the kitchen.
func NewResolver(shipmentClient *client.ShipmentClient) *Resolver {
	return &Resolver{shipmentClient: shipmentClient}
}
