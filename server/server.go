package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	auth "github.com/qvntm/Accord/auth"
	pb "github.com/qvntm/Accord/pb"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
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
	clientCert, err := tls.LoadX509KeyPair("../cert/server-cert.pem", "../cert/server-key.pem")
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

type AccordServer struct {
	authServer      *AuthServer
	authInterceptor *AuthInterceptor
	listener        net.Listener
	mutex           sync.RWMutex
	users           map[string]*User
	channels        map[uint64]*Channel
	jwtManager      *auth.JWTManager
}

func NewAccordServer() *AccordServer {
	authServer := NewAuthServer()
	return &AccordServer{
		authServer:      authServer,
		authInterceptor: NewAuthInterceptor(authServer.JWTManager()),
		users:           map[string]*User{},
		channels:        map[uint64]*Channel{},
		jwtManager:      auth.NewJWTManager(secretKey, tokenDuration),
	}
}

// Load channels from the persistent storage
func (s *AccordServer) LoadChannels() error {
	return fmt.Errorf("Unimplemented!")
}

// Load channels from the persistent storage
func (s *AccordServer) LoadUsers() error {
	return fmt.Errorf("Unimplemented!")
}

func (s *AccordServer) CreateChannel(_ context.Context, req *pb.CreateChannelRequest) (*pb.CreateChannelResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ch := NewChannel(uint64(len(s.channels)), req.GetName(), req.GetIsPublic())
	s.channels[ch.channelID] = ch
	go ch.Listen()

	res := &pb.CreateChannelResponse{}
	log.Printf("New Channel %s created", req.GetName())
	return res, nil
}

func (s *AccordServer) GetChannels(ctx context.Context, req *pb.GetChannelsRequest) (*pb.GetChannelsResponse, error) {
	if ctx.Err() == context.Canceled {
		log.Println("Client has cancelled the request.")
		return nil, status.Errorf(codes.DeadlineExceeded, "client has cancelled the request")
	}

	if req.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username cannot be empty")
	}

	res := &pb.GetChannelsResponse{
		Channels: []*pb.Channel{},
	}
	for _, channel := range s.channels {
		res.Channels = append(res.Channels, &pb.Channel{
			Id:   channel.channelID,
			Name: channel.name,
		})
	}
	return res, nil
}

func (s *AccordServer) Stream(srv pb.Chat_StreamServer) error {
	var channel *Channel = nil
	var username string = ""
	ctx := srv.Context()

	for {
		req, err := srv.Recv()
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		if username == "" {
			username = req.GetUsername()
		} else if n := req.GetUsername(); username != n {
			return status.Errorf(codes.InvalidArgument, "each stream has to use consistent usernames\nhave:%s\nwant:%s\n", n, username)
		}

		if channel == nil {
			channel = s.channels[req.GetChannelId()]
			if channel == nil {
				return status.Errorf(codes.InvalidArgument, "invalid channel ID: %v", err)
			}
			channel.usersToStreams[username] = srv
		} else if id := req.GetChannelId(); channel.channelID != id {
			return status.Errorf(codes.InvalidArgument, "each stream has to use consistent channel IDs\nhave:%d\nwant:%d\n", id, channel.channelID)
		}

		var channelMessage *Message = nil

		// verify StreamRequest contains a valid message type
		switch req.GetMsg().(type) {
		case *pb.StreamRequest_UserMsg:
			// TODO: distinguish between WRITE/MODIFY/DELETE messages.
			msg := req.GetUserMsg()
			channelMessage = &Message{
				timestamp: time.Now(),
				from:      username,
				content:   msg.GetContent(),
			}
		case *pb.StreamRequest_ConfMsg:
			// TODO: Implement configuration message changes.
			return status.Errorf(codes.Unimplemented, "configuration message handling is not implemented")
		default:
			return status.Errorf(codes.InvalidArgument, "invalid message type")
		}

		if channelMessage != nil {
			select {
			// handle abrupt client disconnection
			case <-ctx.Done():
				return status.Error(codes.Canceled, ctx.Err().Error())
			case channel.msgc <- *channelMessage:
			}
		}
	}

	return nil
}

func (s *AccordServer) Listen(serv_addr string) (string, error) {
	listener, err := net.Listen("tcp", serv_addr)
	if err != nil {
		log.Print("Failed to init listener:", err)
		return "", err
	}
	log.Print("Initialized listener:", listener.Addr().String())

	s.listener = listener
	return s.listener.Addr().String(), nil
}

func (s *AccordServer) Start() {
	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("Cannot load TLS credentials:", err)
	}
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(s.authInterceptor.Unary()),
		grpc.StreamInterceptor(s.authInterceptor.Stream()),
	}

	serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))

	srv := grpc.NewServer(serverOptions...)
	pb.RegisterAuthServiceServer(srv, s.authServer)
	pb.RegisterChatServer(srv, s)

	srv.Serve(s.listener)
}
