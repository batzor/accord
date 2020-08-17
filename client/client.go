package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/qvntm/Accord/pb"
)

type AccordClient struct {
	pb.ChatClient
	Token string
}

func NewAccordClient() *AccordClient {
	return &AccordClient{}
}

func (cli *AccordClient) Connect(conn_addr string) error {
	conn, err := grpc.Dial(conn_addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
		return err
	}

	cli.ChatClient = pb.NewChatClient(conn)
	fmt.Println("Successfully started!")
	return nil
}

func (cli *AccordClient) CreateUser(username string, password string) error {
	req := &pb.CreateUserRequest{
		Username: username,
		Password: password,
	}

	log.Print("Creating user...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := cli.ChatClient.CreateUser(ctx, req)
	return err
}

func (cli *AccordClient) Login(username string, password string) error {
	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	log.Print("Logging in...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := cli.ChatClient.Login(ctx, req)
	if err == nil {
		log.Print("Acquired new token: ", cli.Token)
		cli.Token = res.GetToken()
	}

	return err
}
