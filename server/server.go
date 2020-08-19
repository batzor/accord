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
	// adding dummy data so far.
	user1, _ := NewUser("baz", "baz123", "member")
	user2, _ := NewUser("tmr", "tmr456", "member")
	dummyUsersMap := map[string]*User{
		"baz": user1,
		"tmr": user2,
	}
	dummyUsersSlice := []User{*user1, *user2}
	channels := map[uint64]*Channel{
		01234: NewChannel(01234, dummyUsersSlice, false),
		56789: NewChannel(56789, dummyUsersSlice, true),
	}
	return &AccordServer{
		users:      dummyUsersMap, // ideally, has to be empty when server is initialized.
		channels:   channels,
		jwtManager: auth.NewJWTManager(secretKey, tokenDuration),
	}
}

func (s *AccordServer) PrepareChannels() {
	for _, channel := range s.channels {
		go channel.Listen()
	}
}

func (s *AccordServer) AddNewChannel(ch Channel) {
	s.channels[ch.channelID] = &ch
	go ch.Listen()
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

func (s *AccordServer) Start(serv_addr string) {
	fmt.Println("Starting up!")
	listener, err := net.Listen("tcp", serv_addr)
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}

	s.PrepareChannels()

	srv := grpc.NewServer()
	pb.RegisterChatServer(srv, s)

	srv.Serve(listener)
}
