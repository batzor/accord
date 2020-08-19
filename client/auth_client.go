package client

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/qvntm/Accord/pb"
)

// AuthClient is a client to call authentication RPC
type AuthClient struct {
	pb.AuthServiceClient
}

// NewAuthClient returns a new auth client
func NewAuthClient(cc *grpc.ClientConn) *AuthClient {
	return &AuthClient{
		AuthServiceClient: pb.NewAuthServiceClient(cc),
	}
}

func (c *AuthClient) CreateUser(username string, password string) error {
	req := &pb.CreateUserRequest{
		Username: username,
		Password: password,
	}

	log.Print("Creating user...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.AuthServiceClient.CreateUser(ctx, req)
	return err
}

// Login login user and returns the access token
func (c *AuthClient) Login(username string, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	res, err := c.AuthServiceClient.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
