package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Tanmoy095/LogiSynapse/shared/contracts"
	"github.com/Tanmoy095/LogiSynapse/shared/proto"
)

//Now we implement the actual work.
//Best Practice Note: To avoid code duplication, your shipment-service should eventually
// be refactored to use these same functions or libraries. For now,
// we will place the logic here.

type ShipmentActivities struct {
	Store interface {
		CreateShipment(context.Context, contracts.Shipment) (contracts.Shipment, error)
	} // Interface!
	Producer interface {
		Publish(context.Context, string, interface{}) error
	} // Interface!
	ShippoKey string
	Client    *http.Client
}

// Activity 1: The External API Call
func (a *ShipmentActivities) ACTIVITY_CallShippoAPI(ctx context.Context, shipment contracts.Shipment) (contracts.Shipment, error) {
	// Basic validation
	if shipment.Origin == "" || shipment.Destination == "" {
		return contracts.Shipment{}, errors.New("missing required fields")
	}
	// validate package dimenssions
	if shipment.Length <= 0 || shipment.Width <= 0 || shipment.Height <= 0 || shipment.Weight <= 0 || shipment.Unit == "" {
		return contracts.Shipment{}, errors.New("Invalid package Dimensions")

	}
	shippoReq := map[string]interface{}{
		"address_from": map[string]string{
			"city":    shipment.Origin, // e.g., "Dhaka"
			"country": "US",            // Adjust based on origin
		},
		"address_to": map[string]string{
			"city":    shipment.Destination, // e.g., "Berlin"
			"country": "BD",                 // Adjust for destination
		},
		// Dynamic dimensions from mutation
		// Why: Uses client input for accurate shipping, like Amazon
		"parcels": []map[string]interface{}{
			{
				"length": strconv.FormatFloat(shipment.Length, 'f', 2, 64), // e.g., "12.00"
				"width":  strconv.FormatFloat(shipment.Width, 'f', 2, 64),  // e.g., "8.00"
				"height": strconv.FormatFloat(shipment.Height, 'f', 2, 64), // e.g., "1.00"
				"weight": strconv.FormatFloat(shipment.Weight, 'f', 2, 64), // e.g., "0.50"
				"unit":   shipment.Unit,                                    // e.g., "in"
			},
		},
		// Default: let Shippo choose carrier
		// Why: Optimizes cost if client doesn’t specify
		"carrier_account": "",
	}
	if shipment.Carrier.Name != "" {
		carrierMap := map[string]string{
			"FedEx": "fedex",
			"UPS":   "ups",
			"DHL":   "dhl_express",
		}
		if carrierID, ok := carrierMap[shipment.Carrier.Name]; ok {
			shippoReq["carrier_account"] = carrierID
		}

	}
	//convert to JSON
	//Shipoo Api Requires JSON payload
	reqBody, err := json.Marshal(shippoReq)
	if err != nil {
		return contracts.Shipment{}, errors.New("failed to marshal Shippo request: " + err.Error())
	}
	//Create Post request to shippo's /shipment endpoint
	//Book the Shipment
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.goshippo.com/shipments", bytes.NewBuffer(reqBody))
	if err != nil {
		return contracts.Shipment{}, errors.New("failed to create Shippo request: " + err.Error())
	}
	req.Header.Set("Authorization", "ShippoToken "+a.ShippoKey)
	req.Header.Set("Content-Type", "application/json")
	//Send Request
	//Gets Tracking Number status and Url
	resp, err := a.Client.Do(req)
	if err != nil {
		return contracts.Shipment{}, errors.New("failed to call Shippo API: " + err.Error())

	}
	defer resp.Body.Close()
	//Ensure and check Shipment creation Succeeded
	if resp.StatusCode != http.StatusCreated {
		return contracts.Shipment{}, errors.New("Shippo API error: status " + resp.Status)

	}
	//Parse Shippo Response
	//Extract real tracking data
	var shippoResp struct {
		ObjectID       string `json:"object_id"`             // Shippo’s shipment ID
		TrackingNumber string `json:"tracking_number"`       // e.g., "123456789"
		TrackingURL    string `json:"tracking_url_provider"` // e.g., "https://shippo.com/track/123456789"
		Status         string `json:"status"`                // e.g., "PRE_TRANSIT"
		LabelURL       string `json:"label_url"`             // Shipping label URL
		Carrier        string `json:"carrier"`               // e.g., "fedex"
	}
	if err := json.NewDecoder(resp.Body).Decode(&shippoResp); err != nil {
		return contracts.Shipment{}, errors.New("failed to parse Shippo response: " + err.Error())
	}

	//Update response with shippo data
	//Replace static data with shippo Data
	shipment.ID = shippoResp.ObjectID
	shipment.TrackingNumber = shippoResp.TrackingNumber
	// Map Shippo status string to proto enum
	shipment.Status = mapShippoStatusToProto(shippoResp.Status)
	shipment.Carrier = contracts.Carrier{
		Name:        shipment.Carrier.Name,
		TrackingURL: shipment.Carrier.TrackingURL,
	}
	if shipment.Carrier.Name == "" {
		shipment.Carrier.Name = shippoResp.Carrier //User shippos if carrier not provided

	}
	shipment.Length = shipment.Length
	shipment.Width = shipment.Width
	shipment.Height = shipment.Height
	shipment.Weight = shipment.Weight
	shipment.Unit = shipment.Unit

	return shipment, nil
}

// mapShippoStatusToProto maps Shippo status strings to the proto ShipmentStatus enum
func mapShippoStatusToProto(s string) proto.ShipmentStatus {
	switch s {
	case "PRE_TRANSIT":
		return proto.ShipmentStatus_PRE_TRANSIT
	case "IN_TRANSIT":
		return proto.ShipmentStatus_IN_TRANSIT
	case "DELIVERED":
		return proto.ShipmentStatus_DELIVERED
	case "PENDING":
		return proto.ShipmentStatus_PENDING
	case "CANCELLED":
		return proto.ShipmentStatus_CANCELLED
	default:
		return proto.ShipmentStatus_PENDING
	}
}

// Activity 2: The DB Operation
func (a *ShipmentActivities) ACTIVITY_SaveShipmentToDB(ctx context.Context, shipment contracts.Shipment) (contracts.Shipment, error) {
	// Simple wrapper around your existing store logic
	return a.Store.CreateShipment(ctx, shipment)
}

// Activity 3: The Event
func (a *ShipmentActivities) ACTIVITY_PublishKafkaEvent(ctx context.Context, shipment contracts.Shipment) error {
	event := map[string]interface{}{
		"event":   "shipment.created",
		"payload": shipment,
	}
	// Note: We don't use 'go' routine here. Temporal handles the concurrency.
	return a.Producer.Publish(ctx, shipment.ID, event)
}
