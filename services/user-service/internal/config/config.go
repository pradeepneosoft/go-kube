package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCPort    string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		GRPCPort:    getenv("USER_SERVICE_GRPC_PORT", "50051"),
		DatabaseURL: os.Getenv("USER_SERVICE_DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("USER_SERVICE_DATABASE_URL is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
