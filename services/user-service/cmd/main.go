package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	userv1 "github.com/pradeepneosoft/go-kube/gen/proto/user/v1"
	"github.com/pradeepneosoft/go-kube/pkg/migrate"
	"github.com/pradeepneosoft/go-kube/pkg/postgres"
	"github.com/pradeepneosoft/go-kube/services/user-service/internal/config"
	grpcserver "github.com/pradeepneosoft/go-kube/services/user-service/internal/grpc"
	"github.com/pradeepneosoft/go-kube/services/user-service/internal/repository"
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

	migrationPath := filepath.Join("services", "user-service", "migrations", "001_init.sql")
	if err := migrate.Run(ctx, pool, migrationPath); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	repo := repository.NewUserRepository(pool)
	server := grpc.NewServer()
	userv1.RegisterUserServiceServer(server, grpcserver.New(repo))
	reflection.Register(server)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Printf("user-service listening on :%s", cfg.GRPCPort)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("grpc serve stopped: %v", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	server.GracefulStop()
	log.Println("user-service stopped")
}
