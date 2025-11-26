package contracts // <-- Note the package name

import "github.com/Tanmoy095/LogiSynapse/shared/proto"

// Carrier represents a shipping carrier
type Carrier struct {
	Name        string
	TrackingURL string
}

type ShipmentStatus = proto.ShipmentStatus

// Shipment represents the single source of truth for a shipment.
// All internal services (shipment, workflow, etc.) will use this struct.
type Shipment struct {
	ID             string
	Origin         string
	Destination    string
	Eta            string
	Status         proto.ShipmentStatus
	Carrier        Carrier
	TrackingNumber string
	Length         float64
	Width          float64
	Height         float64
	Weight         float64
	Unit           string
}

// ...any other shared models, like Rate...
type Rate struct {
	Carrier       string
	Service       string
	Amount        float64
	EstimatedDays int
}
