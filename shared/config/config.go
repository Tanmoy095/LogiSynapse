// shared/config/config.go
package config

import (
	"fmt"
	"os"
)

// CommonConfig holds infrastructure details used by MULTIPLE services
// We renamed 'Config' to 'CommonConfig'
type CommonConfig struct {
	//Database (PostgreSQL) config
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string
	DB_HOST     string
	DB_PORT     string
	//Kafka config
	KAFKA_TOPIC  string
	KAFKA_BROKER string
	//RabbitMQ config could be added here as well
	RABBITMQ_USER     string
	RABBITMQ_PASSWORD string
	RABBITMQ_HOST     string
	RABBITMQ_PORT     string
}

// LoadCommonConfig returns the shared infrastructure config
// We renamed 'LoadConfig' to 'LoadCommonConfig'
func LoadCommonConfig() *CommonConfig {
	return &CommonConfig{

		DB_USER:     os.Getenv("DB_USER"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_NAME:     os.Getenv("DB_NAME"),

		KAFKA_TOPIC:  os.Getenv("KAFKA_TOPIC"),
		KAFKA_BROKER: os.Getenv("KAFKA_BROKER"),

		RABBITMQ_USER:     os.Getenv("RABBITMQ_USER"),
		RABBITMQ_PASSWORD: os.Getenv("RABBITMQ_PASSWORD"),
		RABBITMQ_HOST:     os.Getenv("RABBITMQ_HOST"),
		RABBITMQ_PORT:     os.Getenv("RABBITMQ_PORT"),
	}
}

// GetDBURL formats the config into a PostgreSQL connection string
func (c *CommonConfig) GetDBURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.DB_USER, c.DB_PASSWORD, c.DB_HOST, c.DB_PORT, c.DB_NAME)
}

// GetRabbitMQURL formats the config into a RabbitMQ connection string

func (c *CommonConfig) GetRabbitMQURL() string {

	//DEFAULTS STANDARD PORTS IF FMISSING PREVENTS CRASHES

	host := c.RABBITMQ_HOST
	if host == "" {
		host = "localhost"
	}
	port := c.RABBITMQ_PORT
	if port == "" {
		port = "5672"
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		c.RABBITMQ_USER, c.RABBITMQ_PASSWORD, host, port)
}
