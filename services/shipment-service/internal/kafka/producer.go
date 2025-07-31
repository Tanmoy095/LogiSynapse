package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

//kafka writter is a component from the kafka go lirbry that handles sending messeges(events) to kafka topic
//writer connects to Kafka and sends messages to a specific topic
//A Kafka broker is a server that runs Kafka and manages the storage and delivery of messages.
//It’s like a post office where messages (events) are stored in topics (mailboxes) and delivered
// to consumers (like status-tracker).

type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer initializes a Kafka writer with the specified broker and topic.
// - brokerURL: Address of the Kafka server (e.g., "localhost:9092").
// - topic: The Kafka topic to publish to (e.g., "shipment.created").
// - Balancer: LeastBytes ensures even distribution across partitions.
func NewKafKaProducer(brokerUrl, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerUrl),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{}, //distribute messeges across platform
		},
	}
}

// Publish sends a message to the Kafka topic.
// - ctx: Context for timeout and cancellation.
// - key: Shipment ID to ensure messages for the same shipment go to the same partition.
// - value: Event payload to be JSON-encoded and sent.

func (c *KafkaProducer) Publish(ctx context.Context, key string, value interface{}) error {
	// Serialize the event payload to JSON bytes for Kafka.
	// JSON is used for compatibility with consumers in different languages.

	bytes, err := json.Marshal(value)
	if err != nil {
		log.Println("❌ Failed to marshal Kafka payload:", err)
		return err
	}
	msg := kafka.Message{
		Key:   []byte(key),
		Value: bytes,
	}
	// send message to  kafka topic via kafka writer
	if err := c.writer.WriteMessages(ctx, msg); err != nil {
		log.Println("❌ Kafka write error:", err)
		return err
	}
	//Log success for debugging and monitoring.
	log.Println("✅ Kafka published:", string(msg.Value))
	return nil

}

// Close shuts down the Kafka writer to free resources.
func (c *KafkaProducer) Close() error {
	return c.writer.Close()
}
