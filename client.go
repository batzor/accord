package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/qvntm/Accord/pb"
)

type AccordClient struct {
	pb.ChatClient
	Token string
}

func NewAccordClient() *AccordClient {
	return &AccordClient{}

}

func (cli *AccordClient) Start() {
	fmt.Println("Starting up!")
	conn, err := grpc.Dial("0.0.0.0:12345", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}

	cli.ChatClient = pb.NewChatClient(conn)
	fmt.Println("Successfully started!")
}

// CreateLaptop calls create laptop RPC
func (cli *AccordClient) CreateUser(username string, password string) {
	req := &pb.CreateUserRequest{
		Username: username,
		Password: password,
	}

	log.Print("Creating user...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := cli.ChatClient.CreateUser(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("Username already exists")
		} else if ok && st.Code() == codes.InvalidArgument {
			log.Print("Invalid password")
		} else {
			log.Fatalf("Unexpected error: %v", err)
		}
		return

	}

	log.Printf("created user: %s", res.Token)
}

func (cli *AccordClient) Login(username string, password string) {
	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	log.Print("Logging in...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := cli.ChatClient.Login(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			log.Printf("Invalid credentials: %v", err)
		} else if ok && st.Code() == codes.Internal {
			log.Fatalf("Server internal error: %v", err)
		}
		return
	}

	cli.Token = res.GetToken()
	log.Print("Acquired new token: ", cli.Token)
	return
}

func main() {
	c := new(AccordClient)
	c.Start()
	c.CreateUser("testuser1", "testpw1")
	c.Login("testuser1", "testpw1")
}
