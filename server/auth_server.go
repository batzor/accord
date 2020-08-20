package server

import (
	"context"
	"log"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	auth "github.com/qvntm/Accord/auth"
	pb "github.com/qvntm/Accord/pb"
)

// AuthServer is the server for authentication
type AuthServer struct {
	mutex      sync.RWMutex
	users      map[string]*User
	jwtManager *auth.JWTManager
}

// NewAuthServer returns a new auth server
func NewAuthServer() *AuthServer {
	return &AuthServer{
		users:      make(map[string]*User),
		jwtManager: auth.NewJWTManager(secretKey, tokenDuration),
	}
}

func (s *AuthServer) JWTManager() *auth.JWTManager {
	return s.jwtManager
}

func (s *AuthServer) GetUser(username string) *User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user := s.users[username]
	if user == nil {
		return nil
	}

	return user.Clone()
}

func (s *AuthServer) CreateUser(_ context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if s.users[req.GetUsername()] != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Username is already in use")
	}

	user, err := NewUser(req.GetUsername(), req.GetPassword(), "")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Password could not be hashed")
	}

	s.users[user.username] = user

	res := &pb.CreateUserResponse{}
	log.Printf("New user %s created", user.username)
	return res, nil
}

func (s *AuthServer) Login(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user := s.GetUser(req.GetUsername())

	if user == nil {
		return nil, status.Errorf(codes.NotFound, "Username not found.")
	}
	if !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect password.")
	}

	token, err := s.jwtManager.Generate(user.username, user.role)
	if err != nil {
		log.Print("token generation failed!")
		return nil, status.Errorf(codes.Internal, "Cannot generate access token")
	}

	res := &pb.LoginResponse{AccessToken: token}
	log.Printf("%s acquired new token", user.username)
	return res, nil
}

func (s *AuthServer) Logout(_ context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Unimplemented!")
}
