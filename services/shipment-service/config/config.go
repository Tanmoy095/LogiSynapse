package config

import (
	"fmt"
	"os"
)

//config holds database connection string

type Config struct {
	DB_USER      string
	DB_PASSWORD  string
	DB_NAME      string
	DB_HOST      string
	DB_PORT      string
	KAFKA_TOPIC  string
	KAFKA_BROKER string
}

//LoadConfig returns a config struct  , it reads environment variable

func LoadConfig() *Config {
	return &Config{
		DB_USER:      os.Getenv("DB_USER"),
		DB_PASSWORD:  os.Getenv("DB_PASSWORD"),
		DB_HOST:      os.Getenv("DB_HOST"),
		DB_PORT:      os.Getenv("DB_PORT"),
		DB_NAME:      os.Getenv("DB_NAME"),
		KAFKA_TOPIC:  os.Getenv("KAFKA_TOPIC"),
		KAFKA_BROKER: os.Getenv("KAFKA_BROKER"),
	}
}

// GetDBURL formats the config into a PostgreSQLconnection string

func (c *Config) GetDBURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.DB_USER, c.DB_PASSWORD, c.DB_HOST, c.DB_PORT, c.DB_NAME)

}
