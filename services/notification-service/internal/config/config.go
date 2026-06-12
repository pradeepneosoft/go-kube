package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL  string
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:  os.Getenv("NOTIFICATION_SERVICE_DATABASE_URL"),
		KafkaTopic:   getenv("KAFKA_ORDERS_TOPIC", "orders.created"),
		KafkaGroupID: getenv("KAFKA_NOTIFICATION_GROUP", "notification-service"),
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	cfg.KafkaBrokers = strings.Split(brokers, ",")

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("NOTIFICATION_SERVICE_DATABASE_URL is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
