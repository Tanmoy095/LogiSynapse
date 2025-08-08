package models

/*
Stores Data afrom ghraphql mutation and shipoo Api for database and Kafkla........
*/

type Shipment struct {
	ID             string //uuid for the shipment
	Origin         string
	Destination    string
	ETA            string  //estimated arrival date
	Status         string  //status like pre_transit,delivered,cancelled
	Carrier        Carrier //Carrier details like traackin url and and carrier name
	TrackingNumber string  //Shipoo's tracking number --> e.g., "1232343543454564
	//package dimensions from mutation for accurate shipment costs ....Basically use for rates ...
	Length float64
	Width  float64
	Height float64
	Weight float64
	Unit   string //unit means inch,cm,m,lb

}

// Carrier Represents carrier details separated for flexibility in  Shipoo Integration
type Carrier struct {
	Name        string //Carrier name like Fedex
	TrackingURL string //e.g-->shipoo.com/track/1233214345434
}
