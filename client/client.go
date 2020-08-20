package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/qvntm/Accord/pb"
)

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile("../cert/ca-cert.pem")
	if err != nil {
		return nil, err

	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")

	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("../cert/client-cert.pem", "../cert/client-key.pem")
	if err != nil {
		return nil, err

	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

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
	fmt.Println("Successfully started!")
	return nil
}

func (c *AccordClient) CreateUser(username string, password string) error {
	return c.authClient.CreateUser(username, password)
}

// CreateChannel sends the request to create new channel.
func (c *AccordClient) CreateChannel(name string, isPublic bool) error {
	if c.ChatClient == nil {
		return fmt.Errorf("Login required")
	}
	req := &pb.CreateChannelRequest{
		Name:     name,
		IsPublic: isPublic,
	}

	log.Print("Creating channel...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.ChatClient.CreateChannel(ctx, req)
	return err
}

func (c *AccordClient) Login(username string, password string) error {
	interceptor, err := NewAuthInterceptor(c.authClient, username, password, 30*time.Second)
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

// GetChannels adds all channels related to the user to the client.
func (c *AccordClient) GetChannels() error {
	req := &pb.GetChannelsRequest{
		Username: c.Username,
		ServerId: c.ServerID,
	}

	log.Println("Getting information about user channels...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	res, err := c.ChatClient.GetChannels(ctx, req)
	if err != nil {
		return fmt.Errorf("Could not get the channel info: %v", err)
	}
	for _, resChannel := range res.GetChannels() {
		c.Channels = append(c.Channels, Channel{
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
func (c *AccordClient) Subscribe(channelID uint64) (*StreamRequestCommunication, *StreamResponseCommunication, error) {
	chatClient, err := c.ChatClient.Stream(context.Background())
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
