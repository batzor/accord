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

func (s *AccordServer) Stream(srv pb.Chat_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "Unimplemented!")
}

func (s *AccordServer) Start(serv_addr string) {
	fmt.Println("Starting up!")
	listener, err := net.Listen("tcp", serv_addr)
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterChatServer(srv, s)

	srv.Serve(listener)
}
