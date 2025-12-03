// workflow-orchestrator/cmd/main.go

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	// 1. Local Imports (Workflow & Activities)
	"github.com/Tanmoy095/LogiSynapse/services/workflow-orchestrator/internal/activities"
	"github.com/Tanmoy095/LogiSynapse/services/workflow-orchestrator/internal/workflow"

	// 2. Shared Infrastructure Imports
	// We use the 'CommonConfig' and 'kafka' from shared
	"github.com/Tanmoy095/LogiSynapse/shared/config"
	pkgkafka "github.com/Tanmoy095/LogiSynapse/shared/kafka"

	// 3. Domain Imports (Reusing Store from shipment-service)
	"github.com/Tanmoy095/LogiSynapse/services/shipment-service/store"
)

func main() {
	// =========================================================================
	// 1. LOAD CONFIG
	// =========================================================================
	// Load the shared infrastructure config (DB creds, Kafka hosts)
	cfg := config.LoadCommonConfig()

	// =========================================================================
	// 2. SETUP DEPENDENCIES (DB & KAFKA)
	// =========================================================================

	//  Connect to Postgres
	// We use cfg.GetDBURL() which comes from your shared/config package
	shipmentStore, err := store.NewPostgresStore(cfg.GetDBURL())
	if err != nil {
		log.Fatalf("Worker failed to connect to DB: %v", err)
	}
	defer shipmentStore.Close()

	//  Connect to Kafka
	// We check if values exist in the config before connecting
	var producer pkgkafka.Publisher

	if cfg.KAFKA_BROKER != "" && cfg.KAFKA_TOPIC != "" {
		producer = pkgkafka.NewKafkaProducer(cfg.KAFKA_BROKER, cfg.KAFKA_TOPIC)
		defer producer.Close()
		log.Println("Worker connected to Kafka")
	} else {
		log.Println("Warning: Kafka config missing, worker will not publish events")
	}

	// =========================================================================
	// 3. SETUP TEMPORAL CLIENT
	// =========================================================================
	// In Docker, we need to connect to "temporal:7233". Locally "localhost:7233".
	temporalHost := os.Getenv("TEMPORAL_HOST_PORT")
	if temporalHost == "" {
		temporalHost = "temporal:7233" // Default for Docker environment
	}

	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()
	log.Println("Worker connected to Temporal at:", temporalHost)

	// =========================================================================
	// 4. REGISTER ACTIVITIES & WORKFLOWS
	// =========================================================================

	// Create the Activity Host and inject the dependencies we just created
	activityHost := &activities.ShipmentActivities{
		Store:     shipmentStore,               // <--- INJECTING THE REAL DB
		Producer:  producer,                    // <--- INJECTING THE REAL KAFKA
		ShippoKey: os.Getenv("SHIPPO_API_KEY"), // Specific env var for this service
		Client:    &http.Client{Timeout: 10 * time.Second},
	}

	// Create the Worker listening to the specific Task Queue
	w := worker.New(c, "SHIPMENT_TASK_QUEUE", worker.Options{})

	// 🚨 CRITICAL FIX: Register Workflow
	// Do NOT call the function with (). Pass the function name only!
	w.RegisterWorkflow(workflow.CreateShimentWorkflow)

	// Register Activities
	w.RegisterActivity(activityHost.ACTIVITY_CallShippoAPI)
	w.RegisterActivity(activityHost.ACTIVITY_SaveShipmentToDB)
	w.RegisterActivity(activityHost.ACTIVITY_PublishKafkaEvent)

	// =========================================================================
	// 5. START WORKER
	// =========================================================================
	log.Println("Worker started. Pollers are running...")

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
