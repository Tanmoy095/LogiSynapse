package workflow

import (
	"time"

	"github.com/Tanmoy095/LogiSynapse/shared/contracts"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// CreateShimentWorkflow is the Clipboard/Blueprint

func CreateShimentWorkflow(ctx workflow.Context, shipment contracts.Shipment) (contracts.Shipment, error) {

	//Configure Retries
	//If shipoo or db is down retry for up to 10 minutes then backoff

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second, // 1 second
		BackoffCoefficient: 2.0,         // Exponential backoff
		MaximumInterval:    time.Minute, // 1 minute
		MaximumAttempts:    100,         // or 0 for infinite retries
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10, // Each step shouldn't take > 10s
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	//Step 1 .. call Shippo API(Activity) to create shipment
	// We pass the raw shipment data, and get back data with a Tracking Number.
	var shippoResult contracts.Shipment

	err := workflow.ExecuteActivity(ctx, "ACTIVITY_CallShippoAPI", shipment).Get(ctx, &shippoResult)
	if err != nil {
		return contracts.Shipment{}, err
	}

	//Step 2: Save to Database (Activity)
	// We save the result from Step 1.

	var storedShipment contracts.Shipment

	err = workflow.ExecuteActivity(ctx, "ACTIVITY_SaveShipmentToDB", shippoResult).Get(ctx, &storedShipment)

	if err != nil {
		return contracts.Shipment{}, err
	}

	//Step 3: Publish Event (Activity)
	// Fire and forget (but Temporal ensures it fires).

	err = workflow.ExecuteActivity(ctx, "Activity_PublishKafkaEvent", storedShipment).Get(ctx, nil)
	if err != nil {
		return contracts.Shipment{}, err

	}

	return storedShipment, nil
}
