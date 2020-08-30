package accord

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/qvntm/accord/pb"
)

// StreamRequestCommunication is used as a communication interface for users
//  of this package who use "Stream" function.
type StreamRequestCommunication struct {
	Reqc   chan<- *ChannelStreamRequest
	Closec <-chan struct{}
}

// StreamResponseCommunication is used as a communication interface for users
//  of this package who use "Stream" function.
type StreamResponseCommunication struct {
	Resc   <-chan *ChannelStreamResponse
	Closec chan<- struct{}
}

type AccordClient struct {
	authClient      *AuthClient
	serverAddr      string
	transportOption grpc.DialOption
	pb.ChatClient
	Username string
	ServerID uint64
	Channels []Channel
}

func NewAccordClient(serverID uint64) *AccordClient {
	return &AccordClient{
		Username: "",
		ServerID: serverID,
	}
}

func (c *AccordClient) AuthClient() *AuthClient {
	return c.authClient
}

func (c *AccordClient) Connect(addr string) error {
	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials:", err)
	}
	c.transportOption = grpc.WithTransportCredentials(tlsCredentials)

	conn, err := grpc.Dial(addr, c.transportOption)
	if err != nil {
		log.Print("Failed to connect to server:", err)
		return err
	}

	c.authClient = NewAuthClient(conn)
	c.serverAddr = addr
	log.Println("Successfully started AuthClient")
	return nil
}

func (c *AccordClient) CreateUser(username string, password string) error {
	return c.authClient.CreateUser(username, password)
}

// CreateChannel sends the request to create new channel.
func (c *AccordClient) CreateChannel(name string, isPublic bool) (uint64, error) {
	if c.ChatClient == nil {
		return 0, fmt.Errorf("Login required")
	}

	req := &pb.AddChannelRequest{
		Name:     name,
		IsPublic: isPublic,
	}
	log.Println("Creating channel...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelID, err := c.ChatClient.AddChannel(ctx, req)
	return channelID.GetChannelId(), err
}

// RemoveChannel permanently deletes the channel and all of its data.
func (c *AccordClient) RemoveChannel(channelID uint64) error {
	if c.ChatClient == nil {
		return fmt.Errorf("Login required")
	}

	req := &pb.RemoveChannelRequest{
		ChannelId: channelID,
	}
	log.Println("Removing the channel...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.ChatClient.RemoveChannel(ctx, req)
	return err
}

func (c *AccordClient) Login(username string, password string) error {
	interceptor, err := NewClientAuthInterceptor(c.authClient, username, password, 30*time.Second)
	if err != nil {
		log.Print("Could not authenticate: ", err)
		return err
	}

	conn, err := grpc.Dial(
		c.serverAddr,
		c.transportOption,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Print("Cannot connect to server: ", err)
		return err
	}

	c.ChatClient = pb.NewChatClient(conn)
	return nil
}

// Subscribe creates stream client and returns communication channels, which are wrapped in structs,
// to which messages can be pushed/received. In the structs, there are also channels for communicating
// when the reqc and resc channels need to be closed.
// "channelID" is only used to check that each request contains the same (consistent) channel ID.
func (c *AccordClient) Subscribe(channelID uint64) (*StreamRequestCommunication, *StreamResponseCommunication, error) {
	chatClient, err := c.ChatClient.ChannelStream(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("ChannelStream RPC failed: %v", err)
	}

	reqc, closereqc := make(chan *ChannelStreamRequest), make(chan struct{})
	go func() {
		defer close(closereqc)
		for {
			msg := <-reqc
			if msg.Username != c.Username {
				log.Printf("Inconsistent usernames in channel: %v\nHave:%s\nWant:%s\n", reqc, msg.Username, c.Username)
				continue
			}
			if msg.ChannelID != channelID {
				log.Printf("Inconsistent channel id used in channel: %v\nHave:%d\nWant:%d\n", reqc, msg.ChannelID, channelID)
				continue
			}
			req := getChannelStreamRequest(msg)
			if err := chatClient.Send(req); err != nil {
				log.Printf("Terminating client stream's send goroutine: %v\n", err)
				return
			}
		}
	}()

	resc, closeresc := make(chan *ChannelStreamResponse), make(chan struct{})
	go func() {
		defer close(resc)
		for {
			res, err := chatClient.Recv()
			if err != nil {
				log.Printf("Terminating client stream's recv goroutine: %v", err)
				return
			}

			resMessage := getChannelStreamResponse(res)
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
