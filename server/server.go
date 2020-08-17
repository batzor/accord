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
	mutex      sync.RWMutex
	users      map[string]*User
	channels   map[uint64]*Channel
	jwtManager *auth.JWTManager
}

func NewAccordServer() *AccordServer {

	return &AccordServer{
		users:      make(map[string]*User),
		jwtManager: auth.NewJWTManager(secretKey, tokenDuration),
	}
}

func (s *AccordServer) PrepareChannels() error {
	// adding dummy channels so far.
	user1, _ := NewUser("baz", "baz123", "member")
	user2, _ := NewUser("tmr", "tmr456", "member")
	dummyUsers := []User{*user1, *user2}
	channels := []Channel{
		*NewChannel(01234, dummyUsers, false),
		*NewChannel(56789, dummyUsers, true),
	}

	for _, channel := range channels {
		go channel.Listen()
	}
	return nil
}

func (s *AccordServer) GetChannelByID(channelID uint64) (*Channel, error) {
	for _, channel := range s.channels {
		if channel.channelID == channelID {
			return channel, nil
		}
	}
	return nil, fmt.Errorf("channel with id %d does not exist on the server", channelID)
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

func (s *AccordServer) CreateUser(_ context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if s.users[req.GetUsername()] != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Username is already in use")
	}

	user, err := NewUser(req.GetUsername(), req.GetPassword(), "")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Password could not be hashed")
	}

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

	res := &pb.GetChannelsResponse{
		CurrentChannel: uint64(67890),
		Channels: []*pb.Channel{
			{
				Id:   uint64(67890),
				Name: "tmrbaz",
			},
			{
				Id:   uint64(12345),
				Name: "baztmr",
			},
		},
	}
	return res, nil
}

func (s *AccordServer) Stream(srv pb.Chat_StreamServer) error {
	var channel *Channel = nil
	ctx := srv.Context()

	for {
		req, err := srv.Recv()

		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		if channel == nil {
			channel, err = s.GetChannelByID(req.GetChannelId())
			if err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid channel ID: %v", err)
			}
			userID := uint64(12345)
			channel.usersToStreams[userID] = srv
		}

		var channelMessage *Message = nil

		// verify StreamRequest contains a valid message type
		switch req.GetMsg().(type) {
		case *pb.StreamRequest_UserMsg:
			msg := req.GetUserMsg()
			channelMessage = &Message{
				timestamp: time.Now(),
				from:      "don't know",
				content:   msg.GetContent(),
			}
		case *pb.StreamRequest_ConfMsg:
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

func (s *AccordServer) Start(serv_addr string) {
	fmt.Println("Starting up!")
	listener, err := net.Listen("tcp", serv_addr)
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}

	if err := s.PrepareChannels(); err != nil {
		log.Fatalf("Failed to startup servers channels: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterChatServer(srv, s)

	srv.Serve(listener)
}
