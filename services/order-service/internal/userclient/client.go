package userclient

import (
	"context"
	"fmt"

	userv1 "github.com/pradeepneosoft/go-kube/gen/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client userv1.UserServiceClient
}

func New(ctx context.Context, target string) (*Client, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial user service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: userv1.NewUserServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) UserExists(ctx context.Context, userID string) (bool, error) {
	_, err := c.client.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
	if err != nil {
		return false, err
	}
	return true, nil
}
