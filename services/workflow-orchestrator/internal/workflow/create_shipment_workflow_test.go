package workflow

import (
	"testing"

	"github.com/Tanmoy095/LogiSynapse/shared/contracts"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

func TestCreateShipmentWorkflow_Success(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.RegisterActivityWithOptions(func(shipment contracts.Shipment) (contracts.Shipment, error) {
		shipment.ID = "shippo-id-1"
		shipment.TrackingNumber = "trk-1"
		return shipment, nil
	}, activity.RegisterOptions{Name: "ACTIVITY_CallShippoAPI"})

	env.RegisterActivityWithOptions(func(shipment contracts.Shipment) (contracts.Shipment, error) {
		return shipment, nil
	}, activity.RegisterOptions{Name: "ACTIVITY_SaveShipmentToDB"})

	env.RegisterActivityWithOptions(func(shipment contracts.Shipment) error {
		return nil
	}, activity.RegisterOptions{Name: "ACTIVITY_PublishKafkaEvent"})

	input := contracts.Shipment{
		Origin:      "Dhaka",
		Destination: "Berlin",
		Length:      10,
		Width:       10,
		Height:      5,
		Weight:      2,
		Unit:        "kg",
	}
	env.ExecuteWorkflow(CreateShipmentWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result contracts.Shipment
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "shippo-id-1", result.ID)
	require.Equal(t, "trk-1", result.TrackingNumber)
}
