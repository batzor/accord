package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/qvntm/Accord/pb"
)

type Channel struct {
	ID   uint64
	Name string
}

type AccordClient struct {
	pb.ChatClient
	Token            string
	currentChannelID uint64
	channels         []Channel
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

func (cli *AccordClient) GetChannelInfo() {
	req := &pb.GetChannelsRequest{
		UserId:   12345,
		ServerId: 67890,
	}

	log.Println("Getting information about user channels...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	res, err := cli.ChatClient.GetChannels(ctx, req)
	if err != nil {
		log.Fatalf("Could not get the channel info: %v", err)
	}
	cli.currentChannelID = res.GetCurrentChannel()
	for _, resChannel := range res.GetChannels() {
		cli.channels = append(cli.channels, Channel{
			ID:   resChannel.GetId(),
			Name: resChannel.GetName(),
		})
	}
}
