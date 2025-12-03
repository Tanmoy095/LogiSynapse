//shipment-service/service/shipment.service.go

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

	"github.com/Tanmoy095/LogiSynapse/services/shipment-service/store"
	"github.com/Tanmoy095/LogiSynapse/shared/contracts"
	"github.com/Tanmoy095/LogiSynapse/shared/proto"
	"go.temporal.io/sdk/client"
)

// ShipmentService handles business logic.
//
//	We removed 'producer', 'httpClient', and 'shippoKey' from this struct.
//
// Why? Because the "heavy lifting" (API calls, DB transactions) is now done by the
// Workflow Worker. This Service is now just a "Request Initiator".
type ShipmentService struct {
	store          store.ShipmentStore
	temporalClient client.Client // <--- NEW: The connection to the Temporal Server
}

// NewShipmentService creates a new service.
// We now pass the Temporal Client instead of the Kafka Producer.
func NewShipmentService(store store.ShipmentStore, temporalClient client.Client) *ShipmentService {
	return &ShipmentService{
		store:          store,
		temporalClient: temporalClient,
	}
}

// CreateShipment is the "Entry Point".
// Instead of doing the work itself, it delegates everything to Temporal.
func (s *ShipmentService) CreateShipment(ctx context.Context, shipment contracts.Shipment) (contracts.Shipment, error) {
	// catch bad data *before* starting a workflow to save resources.
	if shipment.Origin == "" || shipment.Destination == "" {
		return contracts.Shipment{}, errors.New("missing required fields")
	}

	// Define Workflow Options
	// TaskQueue: This MUST match the queue name defined in your Worker (workflow-orchestrator/cmd/main.go).
	// ID: We use the shipment ID (or generate one) to prevent duplicates (Deduping).
	workflowOptions := client.StartWorkflowOptions{
		ID:        "shipment-create-" + shipment.ID,
		TaskQueue: "SHIPMENT_TASK_QUEUE",
	}

	// Execute the Workflow
	//  We use 'ExecuteWorkflow' instead of 'SignalWithStart' because
	// the gRPC client (frontend) is waiting for a response (the Tracking Number).
	// This call sends the inputs to the Temporal Server.
	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "CreateShipmentWorkflow", shipment)
	if err != nil {
		return contracts.Shipment{}, err
	}

	// waits until the Worker finishes all activities
	// (Call Shippo -> Save DB -> Publish Kafka) and returns the final result.
	var result contracts.Shipment
	err = we.Get(ctx, &result)
	if err != nil {
		return contracts.Shipment{}, err
	}

	return result, nil
}

// Updateshipment updates shipment details in DB.
// 🟡 REFACTOR STATUS: PARTIAL
// We kept the DB update because we still have access to 's.store'.
// We DISABLED the Kafka Event because we removed 's.producer'.
// TODO: In Phase 5, we should turn this into a 'Signal' to the workflow.
func (s *ShipmentService) Updateshipment(ctx context.Context, shipment contracts.Shipment) (contracts.Shipment, error) {
	if shipment.ID == "" {
		return contracts.Shipment{}, errors.New("missing shipment id")
	}

	// Validate existence and status (Read-only check)
	current, err := s.store.GetShipment(ctx, shipment.ID)
	if err != nil {
		return contracts.Shipment{}, errors.New("failed to get shipment: " + err.Error())
	}

	if current.Status != proto.ShipmentStatus_PRE_TRANSIT {
		return contracts.Shipment{}, errors.New("can only update pre_Transit shipment")
	}

	// Merge logic (Keep existing values if new ones are empty)
	updatedShipment := contracts.Shipment{
		ID:             shipment.ID,
		Origin:         ifEmpty(shipment.Origin, current.Origin),
		Destination:    ifEmpty(shipment.Destination, current.Destination),
		Eta:            ifEmpty(shipment.Eta, current.Eta),
		Status:         current.Status,
		Carrier:        contracts.Carrier{Name: ifEmpty(shipment.Carrier.Name, current.Carrier.Name), TrackingURL: current.Carrier.TrackingURL},
		TrackingNumber: current.TrackingNumber,
		Length:         ifZero(shipment.Length, current.Length),
		Width:          ifZero(shipment.Width, current.Width),
		Height:         ifZero(shipment.Height, current.Height),
		Weight:         ifZero(shipment.Weight, current.Weight),
		Unit:           ifEmpty(shipment.Unit, current.Unit),
	}
	//Execute Workflow (Worker handles DB Update + Kafka Event)

	workflowOptions := client.StartWorkflowOptions{
		ID:        "shipment-update-" + updatedShipment.ID,
		TaskQueue: "SHIPMENT_TASK_QUEUE",
	}

	// Execute the Workflow
	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "UpdateShipmentWorkflow", updatedShipment)
	if err != nil {
		return contracts.Shipment{}, err

	}
	// 4. Wait for Result
	var result contracts.Shipment
	err = we.Get(ctx, &result)
	return result, err

}

// DeleteShipment cancels a shipment.
// DeleteShipment starts the CancelShipmentWorkflow
// We DISABLED the Shippo API call because we removed 's.httpClient' and 's.shippoKey'.
func (s *ShipmentService) DeleteShipment(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("missing shipment id")
	}
	shipment, err := s.store.GetShipment(ctx, id)
	if err != nil {
		return errors.New("failed to get shipment: " + err.Error())
	}

	if shipment.Status != proto.ShipmentStatus_PRE_TRANSIT {
		return errors.New("can only delete PRE_TRANSIT shipments")
	}
	// Execute Workflow (Worker handles Shippo Void API + DB Update + Kafka Event)
	workflowOptions := client.StartWorkflowOptions{
		ID:        "shipment-cancel-" + id,
		TaskQueue: "SHIPMENT_TASK_QUEUE",
	}

	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "CancelShipmentWorkflow", shipment)
	if err != nil {
		return err
	}

	//Wait for Completion
	return we.Get(ctx, nil)

}

// GetRates fetches carrier rates.
// Since this is a "Read-Only" operation (it doesn't change state), it doesn't strictly *need* a Workflow.
// However, since we removed 's.httpClient' from the struct, we instantiate a NEW client locally here.
// This allows the function to keep working without the struct dependency.
// GetRates fetches carrier rates from Shippo.
// Enables clients to compare shipping options, like Amazon’s checkout.
// Note: Doesn’t use store since it’s an API call.
func (s *ShipmentService) GetRates(ctx context.Context, origin, destination string, length, width, height, weight float64, unit string) ([]contracts.Rate, error) {
	if origin == "" || destination == "" || length <= 0 || width <= 0 || height <= 0 || weight <= 0 || unit == "" {
		return nil, errors.New("invalid rate input")
	}

	// 🟢 NEW: Create a local HTTP client just for this request
	localClient := &http.Client{Timeout: 10 * time.Second}
	// 🟢 NEW: Read the key directly from Env (since it's gone from the struct)
	shippoKey := os.Getenv("SHIPPO_API_KEY")

	rateReq := map[string]interface{}{
		"address_from": map[string]string{"city": origin, "country": "US"},
		"address_to":   map[string]string{"city": destination, "country": "BD"},
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
		return nil, errors.New("marshal error: " + err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.goshippo.com/rates", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.New("req creation error: " + err.Error())
	}
	req.Header.Set("Authorization", "ShippoToken "+shippoKey)
	req.Header.Set("Content-Type", "application/json")

	// Use the LOCAL client
	resp, err := localClient.Do(req)
	if err != nil {
		return nil, errors.New("shippo api error: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("shippo status error: " + resp.Status)
	}

	var rateResp struct {
		Rates []struct {
			Carrier       string `json:"carrier"`
			Service       string `json:"service"`
			Amount        string `json:"amount"`
			EstimatedDays int    `json:"estimated_days"`
		} `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		return nil, errors.New("decode error: " + err.Error())
	}

	rates := make([]contracts.Rate, len(rateResp.Rates))
	for i, r := range rateResp.Rates {
		amount, _ := strconv.ParseFloat(r.Amount, 64)
		rates[i] = contracts.Rate{
			Carrier:       r.Carrier,
			Service:       r.Service,
			Amount:        amount,
			EstimatedDays: r.EstimatedDays,
		}
	}
	return rates, nil
}

// GetShipments just calls the store, so it still works fine.
func (s *ShipmentService) GetShipments(ctx context.Context, origin string, status proto.ShipmentStatus, destination string, limit, offset int32) ([]contracts.Shipment, error) {
	return s.store.GetShipments(ctx, origin, status, destination, limit, offset)
}

// Helper functions remain unchanged
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
