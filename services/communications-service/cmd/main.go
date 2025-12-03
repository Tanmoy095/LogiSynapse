package main

import (
	"log"

	"github.com/Tanmoy095/LogiSynapse/shared/config"
	"github.com/Tanmoy095/LogiSynapse/shared/rabbitmq"
)

const (
	EmailQueue = "email_Jobs"
	SMSQueue   = "sms_jobs"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// Load common configuration
	cfg := config.LoadCommonConfig()

	//connect to RabbitMQ
	log.Printf("Connecting to RabbitMQ at: %s", cfg.RABBITMQ_HOST)

	amqpURL := cfg.GetRabbitMQURL()
	client, err := rabbitmq.NewClient(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

}
