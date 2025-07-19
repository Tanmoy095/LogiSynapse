package models

/*
This model is the kitchen’s recipe card. It’s a clean, reusable definition of a shipment
that both the business logic (service) and gRPC handler can use. It avoids direct dependency
on proto-generated structs, which are specific to gRPC.

The waiter (GraphQL) and kitchen (gRPC service) need to agree on what a “dish” (shipment)
looks like. This model is the standard recipe card they both understand.
*/

type Carrier struct {
	Name        string
	TrackingURL string
}

type Shipment struct {
	ID          string
	Origin      string
	Destination string
	ETA         string
	Status      string
	Carrier     Carrier
}
