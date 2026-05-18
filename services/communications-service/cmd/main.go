//services/communications-service/cmd/main.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Tanmoy095/LogiSynapse/shared/config"
	pkgkafka "github.com/Tanmoy095/LogiSynapse/shared/kafka"
	pkgrabbit "github.com/Tanmoy095/LogiSynapse/shared/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EmailQueue        = "email_Jobs"
	SMSQueue          = "sms_jobs"
	EmailDLQ          = "email_jobs_dlq"
	SMSDLQ            = "sms_jobs_dlq"
	maxDeliveryReties = 3
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// Load common configuration
	cfg := config.LoadCommonConfig()

	//connect to RabbitMQ
	log.Printf("Connecting to RabbitMQ at: %s", cfg.RABBITMQ_HOST)

	amqpURL := cfg.GetRabbitMQURL()
	rabbitClient, err := pkgrabbit.NewClient(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	//we do not defer client.Close() immediately
	//we want to control exactly when it closes during
	//shutdown sequence below

	if err := rabbitClient.CreateQueue(EmailQueue); err != nil {
		log.Fatalf("Failed to create email queue: %v", err)
	}
	if err := rabbitClient.CreateQueue(SMSQueue); err != nil {
		log.Fatalf("Failed to create SMS queue: %v", err)
	}
	if err := rabbitClient.CreateQueue(EmailDLQ); err != nil {
		log.Fatalf("Failed to create email DLQ: %v", err)
	}
	if err := rabbitClient.CreateQueue(SMSDLQ); err != nil {
		log.Fatalf("Failed to create SMS DLQ: %v", err)
	}

	//Connect to Kafka (The News Ticker)
	// We tune our radio to the "shipment.created" channel.
	var kafkaConsumer *pkgkafka.Consumer

	if cfg.KAFKA_BROKER != "" && cfg.KAFKA_TOPIC != "" {
		log.Printf("Connecting to Kafka at: %s, Topic: %s", cfg.KAFKA_BROKER, cfg.KAFKA_TOPIC)
		kafkaConsumer = pkgkafka.NewConsumer(
			[]string{cfg.KAFKA_BROKER},
			cfg.KAFKA_TOPIC,
			"communications-group")
	}

	//ctx is a signal to tell workers to stop
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Start Email Worker
	wg.Add(1)
	// He runs into the (Goroutine), connects to the 'EmailQueue',
	// and stands there WAITING. He is idle right now because the queue is empty.
	go startEmailWorker(ctx, rabbitClient, &wg)

	//start sms worker
	wg.Add(1)
	go startSmsWorker(ctx, rabbitClient, &wg)

	// --- WORKER 3: The Bridge Dispatcher (The Translator) ---

	if kafkaConsumer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// STORY: Before we start listening, we must write the "Instruction Manual".
			// We define this function HERE, but it does NOT run yet.
			// It only runs when data actually arrives from the internet.

			bridgeHandler := func(ctx context.Context, key []byte, value []byte) error {
				// --- SCENARIO: Shipment #500 Created ---
				// 1. Kafka sends us bytes. 'value' looks like: {"event":"shipment.created", ...}
				// 2. We Parse it into a Map so Go can read i
				var event map[string]interface{}
				if err := json.Unmarshal(value, &event); err != nil {
					log.Printf("Bridge Dispatcher:Failed to unmarshal kafka message:%v", err)
					return err
				}
				//  We check the event name.
				eventType, ok := event["event"].(string)
				if !ok {
					log.Println("Bridge Dispatcher:Event type missing or invalid")
					return nil
				}
				// LOGIC: "Is this a shipment?" -> YES.
				if eventType == "shipment.created" {
					log.Printf("🌉 Bridge: Shipment Event Detected! Creating Email Job...")
					//TransLate the Shipment Event into an Email Job
					// We take the "Fact" (Shipment) and turn it into a "Task" (Email Job).
					emailJob := map[string]interface{}{
						"type":    "welcome_email",
						"payload": event["payload"],
					}
					jsonData, err := json.Marshal(emailJob)
					if err != nil {
						log.Printf("Bridge Dispatcher:Failed to marshal email job:%v", err)
						return err
					}
					// 6. HANDOFF:
					// We walk over to the RabbitMQ Kitchen and drop this ticket in the 'EmailQueue'.
					// NOTE: The 'Email Chef' (from Act 2) is watching this queue.
					// He will see this ticket instantly!
					idempotencyKey := string(key)
					if idempotencyKey == "" {
						idempotencyKey = fmt.Sprintf("shipment-%v", event["payload"])
					}
					wrappedEmail, err := wrapJobWithMeta(jsonData, idempotencyKey, 0)
					if err != nil {
						return err
					}
					if err := rabbitClient.Publish(ctx, EmailQueue, wrappedEmail); err != nil {
						log.Printf("Bridge Dispatcher:Failed to publish email job:%v", err)
						return err
					}
					log.Println("Bridge Dispatcher:Email Job Published to RabbitMQ")
					// --- JOB B: SMS Alert (This was missing!) ---
					smsJob := map[string]interface{}{
						"type":    "sms_alert",
						"payload": event["payload"],
					}
					smsBody, _ := json.Marshal(smsJob)

					// Publish to the SMS Queue
					wrappedSMS, err := wrapJobWithMeta(smsBody, idempotencyKey, 0)
					if err != nil {
						return err
					}
					if err := rabbitClient.Publish(ctx, SMSQueue, wrappedSMS); err != nil {
						// Senior Tip: If Email succeeded but SMS failed, do we fail everything?
						// Ideally, yes, so Kafka retries. But we might send duplicate emails.
						// For now, return error to be safe.
						return err
					}
					log.Println("   -> Sent to SMS Queue")
				}
				return nil
			}

			// STORY: Now that we wrote the instructions, we actually START listening.
			// This function connects to the internet and waits.
			// When a message comes, it grabs the data and calls 'bridgeHandler' (above) with it.
			log.Println("🎧 Bridge Listener Started")
			kafkaConsumer.Start(ctx, bridgeHandler)

		}()
	}

	log.Println("Service running. Press Ctrl + c to stop")
	//waiting for stop signal
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	//the main thread is going to sleep on this line until you
	//press ctrl+c or send a terminate signal
	receivedSignal := <-stopSignal
	log.Printf("Received signal: %v. Initiating shutdown...", receivedSignal)

	log.Println("🛑 Closing time...")
	//cancel stop door this tells workers stop accepting new messege
	cancel() // Tell everyone to stop accepting new work
	//wait for workers to finish processing current messege
	wg.Wait()
	//now all workers quit .we can close rabbitmq connection
	if err := rabbitClient.Close(); err != nil {
		log.Fatalf("Failed to clsoe RabbitMQ connection: %v", err)
	}
	if kafkaConsumer != nil {
		kafkaConsumer.Close()
	}
	log.Println("Service shutdown complete. Safe to exit")

}

//worker Logic

func startEmailWorker(ctx context.Context, client *pkgrabbit.RabbitmqClient, wg *sync.WaitGroup) {

	//signOut when the function finiosh

	defer wg.Done()

	msgs, err := client.Consume(EmailQueue)
	if err != nil {
		log.Printf("Email Worker:Failed to start consuming messages:%v", err)
		return
	}
	for {
		select {
		//manager says stop
		case <-ctx.Done():
			log.Println("Email Worker:Received stop signal.Shutting down...")
			return
		// CASE: A new message arrived
		case d, ok := <-msgs:
			if !ok {
				return
			}

			log.Printf("📧 Email Chef: I got a job! Payload: %s", string(d.Body))

			if err := processWithRetryOrDLQ(ctx, client, d, EmailQueue, EmailDLQ); err != nil {
				log.Printf("Email Worker failed: %v", err)
			}
		}
	}
}

func startSmsWorker(ctx context.Context, client *pkgrabbit.RabbitmqClient, wg *sync.WaitGroup) {

	defer wg.Done()
	msg, err := client.Consume(SMSQueue)
	if err != nil {
		log.Printf("SMS Worker:Failed to start consuming messages:%v", err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			log.Println("SMS Worker:Received stop signal.Shutting down...")
			return

		case d, ok := <-msg:
			if !ok {
				return
			}
			log.Printf("📱 Processing SMS: %s", string(d.Body))
			if err := processWithRetryOrDLQ(ctx, client, d, SMSQueue, SMSDLQ); err != nil {
				log.Printf("SMS Worker failed: %v", err)
			}
		}

	}

}

func processWithRetryOrDLQ(ctx context.Context, client *pkgrabbit.RabbitmqClient, d amqp.Delivery, queueName, dlqName string) error {
	idempotencyKey, retryCount, payload, err := unwrapJobWithMeta(d.Body)
	if err != nil {
		return err
	}
	processErr := processJob(payload)
	if processErr == nil {
		if err := d.Ack(false); err != nil {
			return err
		}
		return nil
	}

	if retryCount >= maxDeliveryReties {
		wrapped, wrapErr := wrapJobWithMeta(payload, idempotencyKey, retryCount)
		if wrapErr != nil {
			return wrapErr
		}
		if err := client.Publish(ctx, dlqName, wrapped); err != nil {
			return err
		}
		return d.Ack(false)
	}

	wrapped, wrapErr := wrapJobWithMeta(payload, idempotencyKey, retryCount+1)
	if wrapErr != nil {
		return wrapErr
	}
	if err := client.Publish(ctx, queueName, wrapped); err != nil {
		return err
	}
	return d.Ack(false)
}

func processJob(body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("empty job payload")
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}

type queuedJob struct {
	IdempotencyKey string          `json:"idempotency_key"`
	RetryCount     int32           `json:"retry_count"`
	Payload        json.RawMessage `json:"payload"`
}

func wrapJobWithMeta(body []byte, idempotencyKey string, retryCount int32) ([]byte, error) {
	return json.Marshal(queuedJob{
		IdempotencyKey: idempotencyKey,
		RetryCount:     retryCount,
		Payload:        json.RawMessage(body),
	})
}

func unwrapJobWithMeta(body []byte) (string, int32, []byte, error) {
	var job queuedJob
	if err := json.Unmarshal(body, &job); err != nil {
		return "", 0, nil, err
	}
	return job.IdempotencyKey, job.RetryCount, []byte(job.Payload), nil
}

/*
KafkaConsumer start after producing to rabbitmq? How is it possible?"

It is NOT starting after.

First: kafkaConsumer.Start runs. It connects to the internet and waits.

Second: A Kafka message arrives.

Third: kafkaConsumer looks at the bridgeHandler function you gave it.

Fourth: It runs the code inside that function, which triggers rabbitClient.Publish.

3. Visual Timeline

Time 00:00 (App Start): bridgeHandler is defined (Instructions written).

Time 00:01: kafkaConsumer.Start is called (Chef hired).

Time 00:02 - 00:09: Nothing happens. The app is just listening. Publish has never run.

Time 00:10: Event Arrives! (Shipment Created).

Time 00:11: The kafkaConsumer executes the instructions. NOW Publish runs.
*/
