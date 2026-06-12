package grpcserver

import (
	"context"
	"errors"

	userv1 "github.com/pradeepneosoft/go-kube/gen/proto/user/v1"
	"github.com/pradeepneosoft/go-kube/services/user-service/internal/model"
	"github.com/pradeepneosoft/go-kube/services/user-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	repo *repository.UserRepository
}

func New(repo *repository.UserRepository) *Server {
	return &Server{repo: repo}
}

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.User, error) {
	user, err := s.repo.Create(ctx, req.GetEmail(), req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create user: %v", err)
	}

	return toProto(user), nil
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.User, error) {
	user, err := s.repo.GetByID(ctx, req.GetId())
	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get user: %v", err)
	}

	return toProto(user), nil
}

func (s *Server) ListUsers(ctx context.Context, _ *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %v", err)
	}

	resp := &userv1.ListUsersResponse{
		Users: make([]*userv1.User, 0, len(users)),
	}
	for _, user := range users {
		resp.Users = append(resp.Users, toProto(user))
	}

	return resp, nil
}

func toProto(user model.User) *userv1.User {
	return &userv1.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
