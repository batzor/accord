package client

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
)

type AccordClient struct {
	authClient *AuthClient
}

func NewAccordClient() *AccordClient {
	return &AccordClient{}
}

func (c *AccordClient) AuthClient() *AuthClient {
	return c.authClient
}

func (c *AccordClient) Connect(conn_addr string) error {
	conn, err := grpc.Dial(conn_addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
		return err
	}

	c.authClient = NewAuthClient(conn)
	fmt.Println("Successfully started!")
	return nil
}
