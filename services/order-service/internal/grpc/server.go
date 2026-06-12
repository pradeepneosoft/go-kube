package grpcserver

import (
	"context"
	"errors"
	"log"

	orderv1 "github.com/pradeepneosoft/go-kube/gen/proto/order/v1"
	"github.com/pradeepneosoft/go-kube/pkg/events"
	"github.com/pradeepneosoft/go-kube/pkg/kafka"
	"github.com/pradeepneosoft/go-kube/services/order-service/internal/model"
	"github.com/pradeepneosoft/go-kube/services/order-service/internal/repository"
	"github.com/pradeepneosoft/go-kube/services/order-service/internal/userclient"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	orderv1.UnimplementedOrderServiceServer
	repo       *repository.OrderRepository
	users      *userclient.Client
	producer   *kafka.Producer
	kafkaTopic string
}

func New(
	repo *repository.OrderRepository,
	users *userclient.Client,
	producer *kafka.Producer,
	kafkaTopic string,
) *Server {
	return &Server{
		repo:       repo,
		users:      users,
		producer:   producer,
		kafkaTopic: kafkaTopic,
	}
}

func (s *Server) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.Order, error) {
	exists, err := s.users.UserExists(ctx, req.GetUserId())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Unavailable, "user service unavailable: %v", err)
	}
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	order, err := s.repo.Create(ctx, req.GetUserId(), req.GetProductName(), int(req.GetQuantity()))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create order: %v", err)
	}

	event := events.NewOrderCreated(order.ID, order.UserID, order.ProductName, order.Quantity)
	if err := s.producer.PublishJSON(ctx, order.ID, event); err != nil {
		log.Printf("publish order event failed: %v", err)
		return nil, status.Errorf(codes.Internal, "publish order event: %v", err)
	}

	return toProto(order), nil
}

func (s *Server) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.Order, error) {
	order, err := s.repo.GetByID(ctx, req.GetId())
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get order: %v", err)
	}

	return toProto(order), nil
}

func (s *Server) ListOrdersByUser(ctx context.Context, req *orderv1.ListOrdersByUserRequest) (*orderv1.ListOrdersByUserResponse, error) {
	orders, err := s.repo.ListByUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "list orders: %v", err)
	}

	resp := &orderv1.ListOrdersByUserResponse{
		Orders: make([]*orderv1.Order, 0, len(orders)),
	}
	for _, order := range orders {
		resp.Orders = append(resp.Orders, toProto(order))
	}

	return resp, nil
}

func toProto(order model.Order) *orderv1.Order {
	return &orderv1.Order{
		Id:          order.ID,
		UserId:      order.UserID,
		ProductName: order.ProductName,
		Quantity:    int32(order.Quantity),
		Status:      order.Status,
		CreatedAt:   order.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
