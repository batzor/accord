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
	Channels map[uint64]*ClientChannel
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.ChatClient.RemoveChannel(ctx, req)
	return err
}

func (c *AccordClient) GetChannels() error {
	if c.ChatClient == nil {
		return fmt.Errorf("Login required")
	}

	req := &pb.GetChannelsRequest{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.ChatClient.GetChannels(ctx, req)
	if err != nil {
		return err
	}

	metas := res.GetChannelMetas()
	// Remove non-existing channels
	for k := range c.Channels {
		_, ok := metas[k]
		if !ok {
			delete(c.Channels, k)
		}
	}

	// Update channel metadatas
	for k, meta := range metas {
		c.Channels[k] = NewClientChannel(k, meta.Name, meta.IsPublic)
	}

	return nil
}

func (c *AccordClient) GetChannel(channelID uint64) error {
	if c.ChatClient == nil {
		return fmt.Errorf("Login required")
	}

	req := &pb.GetChannelRequest{
		ChannelId: channelID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.ChatClient.GetChannel(ctx, req)
	if err != nil {
		return err
	}

	data := res.GetChannel()
	c.Channels[channelID].Name = data.GetName()
	c.Channels[channelID].PinnedMsgId = data.GetPinnedMsgId()
	c.Channels[channelID].IsPublic = data.GetIsPublic()

	users := data.GetUsers()

	// Remove non-existing users
	for uname := range c.Channels[channelID].Users {
		_, ok := users[uname]
		if !ok {
			delete(c.Channels[channelID].Users, uname)
		}
	}

	for uname, user := range users {
		c.Channels[channelID].Users[uname] = Role(user.GetRole())
	}

	return nil
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

// Subscribe returns the channel, which will send all the updates about the channel.
func (c *AccordClient) Subscribe(channelID uint64) (*StreamResponseCommunication, error) {
	channel, ok := c.Channels[channelID]
	if !ok {
		return nil, fmt.Errorf("there is no channel with id %d in the server or it has not been fetched yet", channelID)
	}
	if !channel.IsFetched {
		return nil, fmt.Errorf("channel with id %d has not been fetched yet", channelID)
	}

	// TODO: I think this needs to be reorganized.
	// Current state: process one message and then wait until receiver reads it.
	// Desired state: process all messages from server in for loop and then notify
	// the receiver about those asynchronously.
	resc, closeresc := make(chan *ChannelStreamResponse), make(chan struct{})
	go func() {
		defer close(resc)
		for {
			res, err := channel.Stream.Recv()
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

	resComm := &StreamResponseCommunication{
		Resc:   resc,
		Closec: closeresc,
	}
	return resComm, nil
}

func (c *AccordClient) Send(msg *ChannelStreamRequest) error {
	channel, ok := c.Channels[msg.ChannelID]
	if !ok {
		return fmt.Errorf("there is no channel with id %d in the server or it has not been fetched yet", msg.ChannelID)
	}
	if !channel.IsFetched {
		return fmt.Errorf("channel with id %d has not been fetched yet", msg.ChannelID)
	}

	req := getChannelStreamRequest(msg)
	if err := channel.Stream.Send(req); err != nil {
		return fmt.Errorf("Failed to send request %v to the channel stream %v", req, channel.Stream)
	}

	return nil
}
