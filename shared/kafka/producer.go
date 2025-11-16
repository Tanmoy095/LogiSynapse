package kafka

import (
	"context"
	"encoding/json"
	"log"

	skafka "github.com/segmentio/kafka-go"
)

// Writer defines the subset of segmentio kafka.Writer we need. This makes the producer testable.
type Writer interface {
	WriteMessages(ctx context.Context, msgs ...skafka.Message) error
	Close() error
}

// Publisher is the interface used by services to publish events.
type Publisher interface {
	Publish(ctx context.Context, key string, value interface{}) error
	Close() error
}

// KafkaProducer is a thin wrapper around a kafka writer implementing Publisher.
type KafkaProducer struct {
	writer Writer
}

// NewKafkaProducer creates a real KafkaProducer that writes to the provided broker/topic.
func NewKafkaProducer(brokerURL, topic string) *KafkaProducer {
	w := &skafka.Writer{
		Addr:     skafka.TCP(brokerURL),
		Topic:    topic,
		Balancer: &skafka.LeastBytes{},
	}
	return &KafkaProducer{writer: w}
}

// NewKafkaProducerWithWriter allows injecting a test writer.
func NewKafkaProducerWithWriter(w Writer) *KafkaProducer {
	return &KafkaProducer{writer: w}
}

// Publish marshals the value to JSON and writes a kafka message with the given key.
func (p *KafkaProducer) Publish(ctx context.Context, key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		log.Println("failed to marshal kafka value:", err)
		return err
	}
	msg := skafka.Message{Key: []byte(key), Value: b}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Println("kafka write error:", err)
		return err
	}
	return nil
}

// Close closes the underlying writer.
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
