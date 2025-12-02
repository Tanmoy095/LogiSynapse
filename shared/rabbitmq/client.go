package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqClient struct {
	//conn is a tcp connection to rabbitmq server
	conn *amqp.Connection
	chn  *amqp.Channel
}

func NewClient(url string) (*RabbitmqClient, error) {
	//Dial the server
	//this opens the tcp connection to rabbitmq server
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	//Open a channel. This open a logical session inside the connection.
	chn, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel ")
	}

	// Return the ready to use  client

	return &RabbitmqClient{
		conn: conn,
		chn:  chn,
	}, nil

}

// close cleans up
func (r *RabbitmqClient) Close() error {
	if err := r.chn.Close(); err != nil {
		return err
	}
	if err := r.conn.Close(); err != nil {
		return err
	}
	return nil

}

// Crreate Queue prepares a queue to hold messeege
func (r *RabbitmqClient) CreateQueue(queueName string) error {
	_, err := r.chn.QueueDeclare(
		queueName, //name of queue
		true,      //durable
		false,     //delete when unused
		false,     //exclusive
		false,     //no-wait
		nil,       //arguments
	)
	return err
}

// publish sends a messege to a specific queue

func (r *RabbitmqClient) Publish(ctx context.Context, queueName string, body []byte) error {
	return r.chn.PublishWithContext(
		ctx,
		"",        //exchange
		queueName, //routing key (queue name)
		false,     //mandatory
		false,     //immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // make message persistent
			Body:         body,            //actual data payload
		},
	)
}

// Consume start listening for  messege from a specific queue..
//It returns a read only channel
// that delivers message as they arrive

func (r *RabbitmqClient) Consume(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := r.chn.Consume(
		queueName, //queue
		"",        //consumer
		false,     //auto-ack
		false,     //exclusive
		false,     //no-local
		false,     //no-wait
		nil,       //args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
