package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer holds the connection to the Kafka server.
// Think of this as the "Radio Receiver".
type Consumer struct {
	reader *kafka.Reader
}

// Handler is a "Function Type".
// This describes the shape of the "Recipe" that main.go will pass to us.
// We (the Consumer) don't know WHAT the recipe is, we just know how to run it.
type Handler func(ctx context.Context, key []byte, value []byte) error

// NewConsumer creates the connection (The Radio).
// groupID is crucial: If you run 10 copies of this app, the GroupID ensures
// they split the work instead of all 10 processing the same message.
func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID, 
			MinBytes: 10e3,    // 10KB (Wait for a little data pack)
			MaxBytes: 10e6,    // 10MB
		}),
	}
}

// Start begins the "Shift". It is an infinite loop that never stops.
// This is the function you called in main.go with: kafkaConsumer.Start(ctx, bridgeHandler)
func (c *Consumer) Start(ctx context.Context, handler Handler) {
	log.Printf("🎧 Kafka Consumer started. Topic: %s, Group: %s", c.reader.Config().Topic, c.reader.Config().GroupID)

	// THE INFINITE LOOP
	for {
		// 1. Check if the Manager (main.go) cancelled the context (Shutdown)
		if ctx.Err() != nil {
			return // Stop working
		}

		// 2. WAIT for a message (FetchMessage)
		// 🛑 THE CODE PAUSES HERE! 🛑
		// The program sleeps on this line until a message arrives from the internet.
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			// If we are shutting down, just exit
			if ctx.Err() != nil {
				return
			}
			// If it's a network error, log it and wait a bit before trying again
			log.Printf("⚠️ Error fetching message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// 3. EXECUTE THE RECIPE (The Callback)
		// We have the ingredients (m.Key, m.Value).
		// Now we call the function 'handler' that you wrote in main.go.
		// THIS is the moment 'bridgeHandler' and 'rabbitClient.Publish' actually run.
		
		// We give the handler 10 seconds to finish. If it takes longer, we cut it off.
		processCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err = handler(processCtx, m.Key, m.Value) 
		cancel() // Clean up context

		// 4. Did the Recipe fail?
		if err != nil {
			log.Printf("❌ Processing failed (Offset %d): %v", m.Offset, err)
			
			// SENIOR LOGIC:
			// If the handler returned an error (e.g., RabbitMQ was down),
			// we do NOT commit the message.
			// This means Kafka thinks we haven't done it yet.
			// Kafka will send this SAME message to us again in a few seconds (Retry).
			continue 
		}

		// 5. Success! Commit the message.
		// "Kafka, I am done with message #500. You can mark it as read."
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("❌ Failed to commit offset: %v", err)
		}
	}
}

// Close disconnects from the server.
func (c *Consumer) Close() error {
	return c.reader.Close()
}