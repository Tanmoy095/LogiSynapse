package models

/*
This model is the kitchen’s recipe card. It’s a clean, reusable definition of a shipment
that both the business logic (service) and gRPC handler can use. It avoids direct dependency
on proto-generated structs, which are specific to gRPC.

The waiter (GraphQL) and kitchen (gRPC service) need to agree on what a “dish” (shipment)
looks like. This model is the standard recipe card they both understand.
*/

import "github.com/Tanmoy095/LogiSynapse/shared/proto"

// Carrier represents a shipping carrier
type Carrier struct {
	Name        string
	TrackingURL string
}
type ShipmentStatus = proto.ShipmentStatus

// Shipment represents a shipment entity
type Shipment struct {
	ID          string
	Origin      string
	Destination string
	Eta         string
	Status      proto.ShipmentStatus
	Carrier     Carrier
	// Additional fields used by service and store
	TrackingNumber string
	Length         float64
	Width          float64
	Height         float64
	Weight         float64
	Unit           string
}
