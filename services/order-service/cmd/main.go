package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	orderv1 "github.com/pradeepneosoft/go-kube/gen/proto/order/v1"
	"github.com/pradeepneosoft/go-kube/pkg/kafka"
	"github.com/pradeepneosoft/go-kube/pkg/migrate"
	"github.com/pradeepneosoft/go-kube/pkg/postgres"
	"github.com/pradeepneosoft/go-kube/services/order-service/internal/config"
	grpcserver "github.com/pradeepneosoft/go-kube/services/order-service/internal/grpc"
	"github.com/pradeepneosoft/go-kube/services/order-service/internal/repository"
	"github.com/pradeepneosoft/go-kube/services/order-service/internal/userclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	migrationPath := filepath.Join("services", "order-service", "migrations", "001_init.sql")
	if err := migrate.Run(ctx, pool, migrationPath); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	userClient, err := userclient.New(ctx, cfg.UserServiceURL)
	if err != nil {
		log.Fatalf("connect user service: %v", err)
	}
	defer userClient.Close()

	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer producer.Close()

	repo := repository.NewOrderRepository(pool)
	server := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(server, grpcserver.New(repo, userClient, producer, cfg.KafkaTopic))
	reflection.Register(server)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Printf("order-service listening on :%s (user-service=%s kafka=%s topic=%s)",
		cfg.GRPCPort, cfg.UserServiceURL, cfg.KafkaBrokers, cfg.KafkaTopic)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("grpc serve stopped: %v", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	server.GracefulStop()
	log.Println("order-service stopped")
}
