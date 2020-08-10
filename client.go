package main

import (
	"fmt"
	"log"

	"google.golang.org/grpc"

	pb "github.com/qvntm/Accord/proto"
)

type AccordClient struct {
	pb.ChatClient
	Token string
}

func (c *AccordClient) Start() {
	fmt.Println("Starting up!")
	conn, err := grpc.Dial("localhost:12345", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to init listener: %v", err)
	}
	defer conn.Close()

	c.ChatClient = pb.NewChatClient(conn)
	fmt.Println("Successfully started!")
}

func main() {
	c := new(AccordClient)
	c.Start()
}
