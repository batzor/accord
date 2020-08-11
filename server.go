package main

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

	pb "github.com/qvntm/Accord/pb"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

type AccordServer struct {
	mutex      sync.RWMutex
	users      map[string]*User
	jwtManager *JWTManager
}

func NewAccordServer() *AccordServer {

	return &AccordServer{
		users:      make(map[string]*User),
		jwtManager: NewJWTManager(secretKey, tokenDuration),
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

	res := &pb.CreateUserResponse{
		Token: "somesortoftoken",
	}
	log.Printf("New user %s created", user.Username)
	return res, nil
}

func (s *AccordServer) Login(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user := s.GetUser(req.GetUsername())

	if user == nil {
		return nil, status.Errorf(codes.NotFound, "incorrect username")
	}
	if !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect password")
	}

	token, err := s.jwtManager.Generate(user)
	if err != nil {
		log.Print("token generation failed!")
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.LoginResponse{Token: token}
	log.Printf("%s acquired new token", user.Username)
	return res, nil
}

func (s *AccordServer) Logout(_ context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return nil, fmt.Errorf("unimplemented!")
}

func (s *AccordServer) Stream(srv pb.Chat_StreamServer) error {
	return fmt.Errorf("unimplemented!")
}

func (s *AccordServer) Start() error {
	fmt.Println("Starting up!")
	listener, err := net.Listen("tcp", "0.0.0.0:12345")
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterChatServer(srv, s)

	return srv.Serve(listener)
}

func main() {
	s := NewAccordServer()
	s.Start()
}
