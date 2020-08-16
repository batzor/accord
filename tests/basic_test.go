package tests

import (
	"testing"

	client "github.com/qvntm/Accord/client"
	server "github.com/qvntm/Accord/server"
	"github.com/stretchr/testify/require"
)

func TestClientCreateUser(t *testing.T) {
	t.Parallel()

	serverAddr := "0.0.0.0:50051"
	go func() {
		s := server.NewAccordServer()
		s.Start(serverAddr)
		t.Log("Server stopped.")
	}()

	c := client.NewAccordClient()
	c.Connect(serverAddr)

	err := c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw1")
	require.NoError(t, err)
}

func TestClientLogin(t *testing.T) {
	t.Parallel()

	serverAddr := "0.0.0.0:50051"
	go func() {
		s := server.NewAccordServer()
		s.Start(serverAddr)
		t.Log("Server stopped.")
	}()

	c := client.NewAccordClient()
	c.Connect(serverAddr)

	err := c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw2")
	require.NotNil(t, err)

	err = c.Login("testuser2", "testpw1")
	require.NotNil(t, err)

	err = c.Login("testuser1", "testpw1")
	require.Nil(t, err)
}
