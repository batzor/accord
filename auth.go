package accord

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/qvntm/accord/pb"
)

// AuthServer is the server for authentication
type AuthServer struct {
	mutex      sync.RWMutex
	users      map[string]*User
	jwtManager *JWTManager
}

// NewAuthServer returns a new auth server
func NewAuthServer() *AuthServer {
	return &AuthServer{
		users:      make(map[string]*User),
		jwtManager: NewJWTManager(secretKey, tokenDuration),
	}
}

func (s *AuthServer) JWTManager() *JWTManager {
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

	user, err := NewUser(req.GetUsername(), req.GetPassword())
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

	token, err := s.jwtManager.Generate(user.username)
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

// AuthClient is a client to call authentication RPC
type AuthClient struct {
	pb.AuthServiceClient
}

// NewAuthClient returns a new auth client
func NewAuthClient(cc *grpc.ClientConn) *AuthClient {
	return &AuthClient{
		AuthServiceClient: pb.NewAuthServiceClient(cc),
	}
}

func (c *AuthClient) CreateUser(username string, password string) error {
	req := &pb.CreateUserRequest{
		Username: username,
		Password: password,
	}

	log.Print("Creating user...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.AuthServiceClient.CreateUser(ctx, req)
	return err
}

// Login login user and returns the access token
func (c *AuthClient) Login(username string, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	res, err := c.AuthServiceClient.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
