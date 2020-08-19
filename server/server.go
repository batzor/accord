package server

import (
	"fmt"
	"log"
	"net"
	"time"

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
	authServer      *AuthServer
	authInterceptor *AuthInterceptor
}

func NewAccordServer() *AccordServer {
	authServer := NewAuthServer()
	return &AccordServer{
		authServer:      authServer,
		authInterceptor: NewAuthInterceptor(authServer.JWTManager()),
	}
}

func (s *AccordServer) Stream(srv pb.Chat_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "Unimplemented!")
}

func (s *AccordServer) Start(serv_addr string) (string, error) {
	fmt.Println("Starting up!")
	listener, err := net.Listen("tcp", serv_addr)
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterAuthServiceServer(srv, s.authServer)
	pb.RegisterChatServer(srv, s)

	go srv.Serve(listener)
	return listener.Addr().String(), nil
}
