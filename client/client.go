package client

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"google.golang.org/grpc"

	pb "github.com/qvntm/Accord/pb"
)

type Channel struct {
	ID   uint64
	Name string
}

// StreamRequestCommunication is used as a communication interface for users
//  of this package who use "Stream" function.
type StreamRequestCommunication struct {
	Reqc   chan<- RequestMessage
	Closec <-chan struct{}
}

// StreamResponseCommunication is used as a communication interface for users
//  of this package who use "Stream" function.
type StreamResponseCommunication struct {
	Resc   <-chan ResponseMessage
	Closec chan<- struct{}
}

type AccordClient struct {
	pb.ChatClient
	Token    string
	Username string
	ServerID uint64
	Channels []Channel
}

func NewAccordClient(serverID uint64) *AccordClient {
	return &AccordClient{
		Token:    "",
		Username: "",
		ServerID: serverID,
	}
}

func (cli *AccordClient) Connect(conn_addr string) error {
	// TODO: add KeepAliveParams like this: https://github.com/grpc/grpc-go/blob/master/examples/features/keepalive/client/main.go
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

func (cli *AccordClient) CreateChannel(name string, isPublic bool) error {
	req := &pb.CreateChannelRequest{
		Name:     name,
		IsPublic: isPublic,
	}

	log.Print("Creating channel...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := cli.ChatClient.CreateChannel(ctx, req)
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

// GetChannelInfo adds all channel to client, related to the user.
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
	return nil
}

// Subscribe creates stream client and returns communication channels, which are wrapped in structs,
// to which messages can be pushed/received. In the structs, there are also channels for communicating
// when the main request and response channels need to be closed.
// "channelID" is only used to check that each request contains same channel ID.
func (cli *AccordClient) Subscribe(channelID uint64) (*StreamRequestCommunication, *StreamResponseCommunication, error) {
	chatClient, err := cli.ChatClient.Stream(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("Stream RPC failed: %v", err)
	}

	reqc, closereqc := make(chan RequestMessage), make(chan struct{})
	go func() {
		defer close(closereqc)
		for {
			msg := <-reqc
			if msg.ChannelID != channelID {
				log.Printf("Inconsistent channel id used in channel: %v. Ignoring it.\n", reqc)
				continue
			}
			switch msg.GetMsg().(type) {
			case *UserRequestMessage:
				userMsg := msg.GetUserMsg()
				req := &pb.StreamRequest{
					Username:  msg.Username,
					ChannelId: msg.ChannelID,
					Msg: &pb.StreamRequest_UserMsg{
						UserMsg: &pb.StreamRequest_UserMessage{
							Type:    UserRequestToPBMessages[userMsg.MsgType],
							Content: userMsg.Content,
						},
					},
				}
				if err := chatClient.Send(req); err != nil {
					log.Printf("Terminating client stream's send goroutine: %v\n", err)
					return
				}
			case *ConfRequestMessage:
				confMsg := msg.GetConfMsg()
				req := &pb.StreamRequest{
					Username:  msg.Username,
					ChannelId: msg.ChannelID,
					Msg: &pb.StreamRequest_ConfMsg{
						ConfMsg: &pb.StreamRequest_ConfMessage{
							Type:        ConfRequestToPBMessages[confMsg.MsgType],
							Placeholder: confMsg.Placeholder,
						},
					},
				}
				if err := chatClient.Send(req); err != nil {
					log.Printf("Terminating client stream's send goroutine: %v\n", err)
					return
				}
			default:
				log.Printf("Invalid message type was passed: %v. Ignoring it.\n", reflect.TypeOf(msg.GetMsg()))
				continue
			}
		}
	}()

	resc, closeresc := make(chan ResponseMessage), make(chan struct{})
	go func() {
		defer close(resc)
		for {
			req, err := chatClient.Recv()
			if err != nil {
				log.Printf("Terminating client stream's recv goroutine: %v", err)
				return
			}

			var resMessage ResponseMessage
			switch req.GetEvent().(type) {
			case *pb.StreamResponse_NewMsg:
				newMsg := req.GetNewMsg()
				resMessage = ResponseMessage{
					Timestamp: req.Timestamp.AsTime(),
					Msg: &NewMessageResponseMessage{
						SenderID: newMsg.SenderId,
						Content:  newMsg.Content,
					},
				}
			case *pb.StreamResponse_UpdateMsg:
				updateMsg := req.GetUpdateMsg()
				resMessage = ResponseMessage{
					Timestamp: req.Timestamp.AsTime(),
					Msg: &UpdateMessageResponseMessage{
						Placeholder: updateMsg.Placeholder,
					},
				}
			default:
				log.Printf("Invalid message type was received: %v. Ignoring it.\n", reflect.TypeOf(req.GetEvent()))
				continue
			}

			select {
			case <-closeresc:
				log.Println("Terminating client stream's send goroutine by the signal of receiver.")
				return
			case resc <- resMessage:
			}
		}
	}()

	reqComm := &StreamRequestCommunication{
		Reqc:   reqc,
		Closec: closereqc,
	}
	resComm := &StreamResponseCommunication{
		Resc:   resc,
		Closec: closeresc,
	}
	return reqComm, resComm, nil
}
