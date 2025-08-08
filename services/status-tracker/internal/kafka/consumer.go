package kafka

// import (
// 	"context"
// 	"log"
// 	"sync"

// 	"github.com/segmentio/kafka-go"
// )

// // ShipmentCreatedEvent represents the structure of the Kafka event from shipment-service.
// // It matches the map[string]interface{} payload sent by the producer.

// type Consumer struct {
// 	reader     *kafka.Reader
// 	trackerSvc *tracker.Service
// 	logger     *log.Logger
// 	topic      string
// }

// // New consumer initialize  the kafka reader with brokewr topic and group id
// func NewKafkaConsumer(broker, topic, groupID string)

// func (c *Consumer) ShipmentCreatedEvent(ctx context.Context, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	c.logger.Printf("Starting kafka consumer on topic: %s", c.topic)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			c.logger.Println("consumer cancelled due to context cancelletion")
// 			if err := c.reader.Close(); err != nil {
// 				c.logger.Printf("Error closing  kafka reader: %v ", err)

// 			}
// 			return
// 		default:
// 			msg, err := c.reader.ReadMessage(ctx)
// 			if err != nil {

// 			}
// 			var event struct{
// 				1nhnnjin
// 			}

// 		}
// 	}
// }
