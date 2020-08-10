package main

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/qvntm/Accord/proto"
)

type server struct{}

func (s *server) Login(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, fmt.Errorf("unimplemented!")
}

func (s *server) Logout(_ context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return nil, fmt.Errorf("unimplemented!")
}

func (s *server) Stream(srv pb.Chat_StreamServer) error {
	return fmt.Errorf("unimplemented!")
}

func main() {
	fmt.Println("Starting up!")
	listener, err := net.Listen("tcp", "0.0.0.0:12345")
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterChatServer(s, &server{})

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to init server: %v", err)
	}
}
