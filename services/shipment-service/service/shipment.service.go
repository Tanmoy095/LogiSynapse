// service/shipment.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	pkgkafka "github.com/Tanmoy095/LogiSynapse/pkg/kafka"
	"github.com/Tanmoy095/LogiSynapse/shared/proto"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
	"github.com/Tanmoy095/LogiSynapse/shipment-service/store"
)

// ShipmentService handles business logic for shipments, using a ShipmentStore for data access.
// Analogy: The chef who uses the pantry (store) to prepare dishes (shipments).
type ShipmentService struct {
	store      store.ShipmentStore // Use interface instead of concrete MemoryStore
	producer   pkgkafka.Publisher
	shippoKey  string
	httpClient *http.Client
}

// NewShipmentService creates a new service with the given store.
// Analogy: Hires a chef and gives them access to the pantry's menu (interface).
func NewShipmentService(store store.ShipmentStore, producer pkgkafka.Publisher) *ShipmentService {
	return &ShipmentService{
		store:     store,
		producer:  producer,
		shippoKey: os.Getenv("SHIPOO_API_KEY"),
		//http client with timeout context

		//Timeout Prevents hanging api calls
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// CreateShipment validates and stores a new shipment.
func (s *ShipmentService) CreateShipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error) {
	// Basic validation
	if shipment.Origin == "" || shipment.Destination == "" {
		return models.Shipment{}, errors.New("missing required fields")
	}
	// validate package dimenssions
	if shipment.Length <= 0 || shipment.Width <= 0 || shipment.Height <= 0 || shipment.Weight <= 0 || shipment.Unit == "" {
		return models.Shipment{}, errors.New("Invalid package Dimensions")

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
		return models.Shipment{}, errors.New("failed to marshal Shippo request: " + err.Error())
	}
	//Create Post request to shippo's /shipment endpoint
	//Book the Shipment
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.goshippo.com/shipments", bytes.NewBuffer(reqBody))
	if err != nil {
		return models.Shipment{}, errors.New("failed to create Shippo request: " + err.Error())
	}
	req.Header.Set("Authorization", "ShippoToken "+s.shippoKey)
	req.Header.Set("Content-Type", "application/json")
	//Send Request
	//Gets Tracking Number status and Url
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return models.Shipment{}, errors.New("failed to call Shippo API: " + err.Error())

	}
	//Ensure and check Shipment creation Succeeded
	if resp.StatusCode != http.StatusCreated {
		return models.Shipment{}, errors.New("Shippo API error: status " + resp.Status)

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
		return models.Shipment{}, errors.New("failed to parse Shippo response: " + err.Error())
	}

	//Update response with shippo data
	//Replace static data with shippo Data
	shipment.ID = shippoResp.ObjectID
	shipment.TrackingNumber = shippoResp.TrackingNumber
	// Map Shippo status string to proto enum
	shipment.Status = mapShippoStatusToProto(shippoResp.Status)
	shipment.Carrier = models.Carrier{
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

	// Store the shipment using the interface
	//store it to postgres
	created, err := s.store.CreateShipment(ctx, shipment)
	// Create a Kafka event payload with the event type and shipment details.
	// Using map[string]interface{} for flexibility in event structure.
	//publish shipment.created event
	//notifies status-Tracker service for real time update
	event := map[string]interface{}{
		"event":   "shipment.created", // Identifies the event type.
		"payload": created,            // Includes shipment details (ID, Origin, Destination, etc.).
	}
	// Publish the event to Kafka in a goroutine for fire-and-forget.
	// Uses the shipment ID as the key to ensure ordered processing in partitions.
	if s.producer != nil {
		go s.producer.Publish(context.Background(), created.ID, event)
	}

	// Return the created shipment for the gRPC response.
	return created, nil
}

//Update_Shipment updates shipment details for pre_Transit shipments

// / GetShipments retrieves shipments based on filters and pagination.
// - origin, status, destination: Filter criteria.
// - limit, offset: Pagination parameters.

// Update shipment
// Update shipment update shipment details for --> pre transit shipments
func (s *ShipmentService) Updateshipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error) {
	//validate shipment id and fields
	//ensure valid update request
	if shipment.ID == "" {
		return models.Shipment{}, errors.New("missing shipment id")

	}
	if shipment.Origin == "" && shipment.Destination == "" && shipment.Eta == "" && shipment.Length == 0 {

		return models.Shipment{}, errors.New("no fields to update")

	}
	//Get existing shipment
	current, err := s.store.GetShipment(ctx, shipment.ID)
	if err != nil {
		return models.Shipment{}, errors.New("no failed to get shipments: " + err.Error())

	}

	//Prevents update if not pre_transit

	if current.Status != proto.ShipmentStatus_PRE_TRANSIT {
		return models.Shipment{}, errors.New("can only update pre_Transit shipment")

	}
	//Marge updated fields
	//preserves existing data for unchanged fields
	updatedShipment := models.Shipment{
		ID:             shipment.ID,
		Origin:         ifEmpty(shipment.Origin, current.Origin),
		Destination:    ifEmpty(shipment.Destination, current.Destination),
		Eta:            ifEmpty(shipment.Eta, current.Eta),
		Status:         current.Status,
		Carrier:        models.Carrier{Name: ifEmpty(shipment.Carrier.Name, current.Carrier.Name), TrackingURL: current.Carrier.TrackingURL},
		TrackingNumber: current.TrackingNumber,
		Length:         ifZero(shipment.Length, current.Length),
		Width:          ifZero(shipment.Width, current.Width),
		Height:         ifZero(shipment.Height, current.Height),
		Weight:         ifZero(shipment.Weight, current.Weight),
		Unit:           ifEmpty(shipment.Unit, current.Unit),
	}
	//Update in postgres
	if err := s.store.UpdateShipment(ctx, updatedShipment); err != nil {
		return models.Shipment{}, err
	}
	//Publish shipment.updated event
	//Notify status tacker of change
	event := map[string]interface{}{
		"event":   "shipment.updated",
		"payload": updatedShipment,
	}
	if s.producer != nil {
		go s.producer.Publish(ctx, updatedShipment.ID, event)
	}
	return updatedShipment, nil

}

// Delete shipment cancel a pre-transit shipment and voids it in shippo
// use updateSHipment to set CANCELLED status instead of separate store method
func (s *ShipmentService) DeleteShipment(ctx context.Context, id string) error {
	// Get shipment to check status
	// Why: Only PRE_TRANSIT shipments can be canceled
	shipment, err := s.store.GetShipment(ctx, id)
	if err != nil {
		return errors.New("failed to get shipment: " + err.Error())
	}

	// Prevent deletion if not PRE_TRANSIT
	// Why: Matches real-world logistics restrictions
	if shipment.Status != proto.ShipmentStatus_PRE_TRANSIT {
		return errors.New("can only delete PRE_TRANSIT shipments")
	}

	// Call Shippo to void the shipment
	// Why: Cancels in Shippo to prevent charges (simplified; assumes id is transaction ID)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.goshippo.com/transactions/"+id+"/void", nil)
	if err != nil {
		return errors.New("failed to create Shippo void request: " + err.Error())
	}
	req.Header.Set("Authorization", "ShippoToken "+s.shippoKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.New("failed to call Shippo void API: " + err.Error())
	}
	defer resp.Body.Close()

	// Check for success
	// Why: Ensures cancellation succeeded
	if resp.StatusCode != http.StatusOK {
		return errors.New("Shippo void API error: status " + resp.Status)
	}

	// Update status to CANCELLED
	// Why: Marks shipment as canceled in database using UpdateShipment
	shipment.Status = proto.ShipmentStatus_CANCELLED
	if err := s.store.UpdateShipment(ctx, shipment); err != nil {
		return err
	}

	// Publish shipment.cancelled event
	// Why: Notifies status-tracker
	event := map[string]interface{}{
		"event":   "shipment.cancelled",
		"payload": shipment,
	}
	if s.producer != nil {
		go s.producer.Publish(ctx, id, event)
	}

	return nil
}

// GetRates fetches carrier rates from Shippo.
// Why: Enables clients to compare shipping options, like Amazon’s checkout.
// Note: Doesn’t use store since it’s an API call.

func (s *ShipmentService) GetShipments(origin string, status proto.ShipmentStatus, destination string, limit, offset int32) ([]models.Shipment, error) {
	return s.store.GetShipments(context.Background(), origin, status, destination, limit, offset)
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

// GetRates fetches carrier rates from Shippo.
// Why: Enables clients to compare shipping options, like Amazon’s checkout.
// Note: Doesn’t use store since it’s an API call.
func (s *ShipmentService) GetRates(ctx context.Context, origin, destination string, length, width, height, weight float64, unit string) ([]models.Rate, error) {
	// Validate input
	// Why: Ensures Shippo gets valid data
	if origin == "" || destination == "" || length <= 0 || width <= 0 || height <= 0 || weight <= 0 || unit == "" {
		return nil, errors.New("invalid rate input")
	}
	// Prepare Shippo rate request
	// Why: Shippo needs address and parcel details for rates
	rateReq := map[string]interface{}{
		"address_from": map[string]string{
			"city":    origin,
			"country": "US",
		},
		"address_to": map[string]string{
			"city":    destination,
			"country": "BD",
		},
		"parcels": []map[string]interface{}{
			{
				"length": strconv.FormatFloat(length, 'f', 2, 64),
				"width":  strconv.FormatFloat(width, 'f', 2, 64),
				"height": strconv.FormatFloat(height, 'f', 2, 64),
				"weight": strconv.FormatFloat(weight, 'f', 2, 64),
				"unit":   unit,
			},
		},
	}
	reqBody, err := json.Marshal(rateReq)
	if err != nil {
		return nil, errors.New("failed to marshal Shippo rate request: " + err.Error())
	}

	// Create POST request to Shippo’s /rates endpoint
	// Why: Fetches available rates
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.goshippo.com/rates", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.New("failed to create Shippo rate request: " + err.Error())
	}
	req.Header.Set("Authorization", "ShippoToken "+s.shippoKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	// Why: Gets rate options
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, errors.New("failed to call Shippo rate API: " + err.Error())
	}
	defer resp.Body.Close()

	// Check for success
	// Why: Ensures rates were fetched
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Shippo rate API error: status " + resp.Status)
	}
	// Parse response
	// Why: Extracts carrier rates
	var rateResp struct {
		Rates []struct {
			Carrier       string `json:"carrier"`
			Service       string `json:"service"`
			Amount        string `json:"amount"`
			EstimatedDays int    `json:"estimated_days"`
		} `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		return nil, errors.New("failed to parse Shippo rate response: " + err.Error())
	}
	// Convert to internal Rate model
	// Why: Prepares rates for gRPC response
	rates := make([]models.Rate, len(rateResp.Rates))
	for i, r := range rateResp.Rates {
		amount, _ := strconv.ParseFloat(r.Amount, 64)
		rates[i] = models.Rate{
			Carrier:       r.Carrier,
			Service:       r.Service,
			Amount:        amount,
			EstimatedDays: r.EstimatedDays,
		}

	}
	return rates, nil
}

func ifEmpty(newValue, oldValue string) string {
	if newValue != "" {
		return newValue

	}
	return oldValue
}
func ifZero(newValue, oldValue float64) float64 {

	if newValue != 0 {
		return newValue

	}
	return oldValue
}
