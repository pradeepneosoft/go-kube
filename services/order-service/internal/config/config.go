package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	GRPCPort       string
	DatabaseURL    string
	UserServiceURL string
	KafkaBrokers   []string
	KafkaTopic     string
}

func Load() (Config, error) {
	cfg := Config{
		GRPCPort:       getenv("ORDER_SERVICE_GRPC_PORT", "50052"),
		DatabaseURL:    os.Getenv("ORDER_SERVICE_DATABASE_URL"),
		UserServiceURL: getenv("USER_SERVICE_URL", "localhost:50051"),
		KafkaTopic:     getenv("KAFKA_ORDERS_TOPIC", "orders.created"),
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	cfg.KafkaBrokers = strings.Split(brokers, ",")

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("ORDER_SERVICE_DATABASE_URL is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
