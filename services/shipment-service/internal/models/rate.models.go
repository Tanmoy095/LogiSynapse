package models

//Rate represents shipping Rate from a carrier
//stores Rate data from shipoos /rate endpoint for client comparison

type Rate struct {
	Carrier       string  //ex:FrdEx
	Service       string  // Service Type ..ex-- FedEx Ground
	Amount        float64 //Cost in USD 15.99 or ...
	EstimatedDays int     //Delivery time in days
}
