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
	Username         string
	ServerID         uint64
	CurrentChannelID uint64
	Channels         []Channel
}

func NewAccordClient(serverID uint64) *AccordClient {
	return &AccordClient{
		Token:    "",
		Username: "",
		ServerID: serverID,
	}
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
		cli.Username = username
	}

	return err
}

func (cli *AccordClient) GetChannelInfo() error {
	if cli.Token == "" && cli.Username == "" {
		return fmt.Errorf("not logged in, JWT token is not obtained or username was not provided")
	}

	req := &pb.GetChannelsRequest{
		Username: cli.Username,
		ServerId: cli.ServerID,
	}

	log.Println("Getting information about user channels...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	res, err := cli.ChatClient.GetChannels(ctx, req)
	if err != nil {
		return fmt.Errorf("Could not get the channel info: %v", err)
	}
	for _, resChannel := range res.GetChannels() {
		cli.Channels = append(cli.Channels, Channel{
			ID:   resChannel.GetId(),
			Name: resChannel.GetName(),
		})
	}
	if len(cli.Channels) > 0 {
		cli.CurrentChannelID = cli.Channels[0].ID
	}
	return nil
}

func (cli *AccordClient) Stream() (<-chan Message, chan<- Message, error) {
	chatClient, err := cli.ChatClient.Stream(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("Stream RPC failed: %v", err)
	}

	sendc := make(chan Message)
	go func() {
		select {
		case msg := <-sendc:
			switch msg.GetMsg().(type) {
			case *UserMessage:
				userMsg := msg.GetUserMsg()
				req := &pb.StreamRequest{
					Msg: &pb.StreamRequest_UserMsg{
						UserMsg: &pb.StreamRequest_UserMessage{
							Type:    UserToPBMessages[userMsg.Type],
							Content: userMsg.Content,
						},
					},
				}
				if err := chatClient.Send(req); err != nil {
					// TODO: Close sendc (Find the appropriate way to do it)
					return
				}
			// TODO: case ConfMessage: (ConfMessage is not declared yet)
			default:
				// TODO: Close sendc (Find the appropriate way to do it)
				return
			}
			// TODO: case <-ctx.Done() (Need smth like that to check for server failure)
		}
	}()

	recvc := make(chan Message)
	/*go func() {
		for {
			req, err := chatClient.Recv()
			if err != nil {

			}
		}
	}()*/

	return recvc, sendc, nil
}
