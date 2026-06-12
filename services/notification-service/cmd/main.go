package main

import (
	"context"
	"log"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pradeepneosoft/go-kube/pkg/kafka"
	"github.com/pradeepneosoft/go-kube/pkg/migrate"
	"github.com/pradeepneosoft/go-kube/pkg/postgres"
	"github.com/pradeepneosoft/go-kube/services/notification-service/internal/config"
	"github.com/pradeepneosoft/go-kube/services/notification-service/internal/handler"
	"github.com/pradeepneosoft/go-kube/services/notification-service/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	migrationPath := filepath.Join("services", "notification-service", "migrations", "001_init.sql")
	if err := migrate.Run(ctx, pool, migrationPath); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	repo := repository.NewNotificationRepository(pool)
	orderHandler := handler.NewOrderCreatedHandler(repo)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer consumer.Close()

	log.Printf("notification-service consuming topic=%s group=%s brokers=%v",
		cfg.KafkaTopic, cfg.KafkaGroupID, cfg.KafkaBrokers)

	if err := consumer.Run(ctx, orderHandler.Handle); err != nil && ctx.Err() == nil {
		log.Fatalf("consumer stopped: %v", err)
	}

	log.Println("notification-service stopped")
}
