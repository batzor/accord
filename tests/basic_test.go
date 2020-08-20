package tests

import (
	"log"
	"testing"

	client "github.com/qvntm/Accord/client"
	server "github.com/qvntm/Accord/server"
	"github.com/stretchr/testify/require"
)

func TestClientCreateUser(t *testing.T) {
	t.Parallel()

	s := server.NewAccordServer()
	addr, err := s.Start("localhost:0")
	log.Print(addr)
	require.NoError(t, err)

	c := client.NewAccordClient()
	c.Connect(addr)

	err = c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw1")
	require.NoError(t, err)
}

func TestClientLogin(t *testing.T) {
	t.Parallel()

	s := server.NewAccordServer()
	addr, err := s.Start("localhost:0")
	log.Print(addr)

	c := client.NewAccordClient()
	c.Connect(addr)

	err = c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw2")
	require.NotNil(t, err)

	err = c.Login("testuser2", "testpw1")
	require.NotNil(t, err)

	err = c.Login("testuser1", "testpw1")
	require.Nil(t, err)
}
