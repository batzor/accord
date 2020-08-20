package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	auth "github.com/qvntm/Accord/auth"
	pb "github.com/qvntm/Accord/pb"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

type AccordServer struct {
	listener   net.Listener
	mutex      sync.RWMutex
	users      map[string]*User
	channels   map[uint64]*Channel
	jwtManager *auth.JWTManager
}

func NewAccordServer() *AccordServer {
	return &AccordServer{
		users:      map[string]*User{},
		channels:   map[uint64]*Channel{},
		jwtManager: auth.NewJWTManager(secretKey, tokenDuration),
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

func (s *AccordServer) GetUser(username string) *User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user := s.users[username]
	if user == nil {
		return nil
	}

	return user.Clone()
}

func (s *AccordServer) CreateChannel(_ context.Context, req *pb.CreateChannelRequest) (*pb.CreateChannelResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ch := NewChannel(uint64(len(s.channels)), req.GetName(), req.GetIsPublic())
	s.channels[ch.channelId] = ch
	go ch.Listen()

	res := &pb.CreateChannelResponse{}
	log.Printf("New Channel %s created", req.GetName())
	return res, nil
}

func (s *AccordServer) CreateUser(_ context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if s.users[req.GetUsername()] != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Username is already in use")
	}

	user, err := NewUser(req.GetUsername(), req.GetPassword(), "")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Password could not be hashed")
	}

	log.Printf("New user %s created", user.Username)
	s.users[user.Username] = user

	res := &pb.CreateUserResponse{}
	log.Printf("New user %s created", user.Username)
	return res, nil
}

func (s *AccordServer) Login(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user := s.GetUser(req.GetUsername())

	if user == nil {
		return nil, status.Errorf(codes.NotFound, "Username not found.")
	}
	if !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect password.")
	}

	token, err := s.jwtManager.Generate(user.Username, user.Role)
	if err != nil {
		log.Print("token generation failed!")
		return nil, status.Errorf(codes.Internal, "Cannot generate access token")
	}

	res := &pb.LoginResponse{Token: token}
	log.Printf("%s acquired new token", user.Username)
	return res, nil
}

func (s *AccordServer) Logout(_ context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Unimplemented!")
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
			Id:   channel.channelId,
			Name: channel.name,
		})
	}
	return res, nil
}

func (s *AccordServer) Stream(srv pb.Chat_StreamServer) error {
	var channel *Channel = nil
	var username string
	ctx := srv.Context()

	for {
		req, err := srv.Recv()
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		if channel == nil {
			channel = s.channels[req.GetChannelId()]
			if channel != nil {
				return status.Errorf(codes.InvalidArgument, "invalid channel ID: %v", err)
			}
			username = "tmr" // TODO: decide how to get username from connection
			channel.usersToStreams[username] = srv
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
	srv := grpc.NewServer()
	pb.RegisterChatServer(srv, s)

	srv.Serve(s.listener)
}
