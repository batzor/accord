package main

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/qvntm/Accord/proto"
)

type AccordServer struct{}

func (s AccordServer) Login(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, fmt.Errorf("unimplemented!")
}

func (s AccordServer) Logout(_ context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return nil, fmt.Errorf("unimplemented!")
}

func (s AccordServer) Stream(srv pb.Chat_StreamServer) error {
	return fmt.Errorf("unimplemented!")
}

func (s AccordServer) Start() error {
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
	s := new(AccordServer)
	s.Start()
}
