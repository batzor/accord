package tests

import (
	"testing"

	"github.com/qvntm/accord"
	"github.com/stretchr/testify/require"
)

func TestClientCreateUser(t *testing.T) {
	t.Parallel()

	serverID := uint64(12345)
	s := accord.NewAccordServer()
	serverAddr, err := s.Listen("localhost:0")
	go func() {
		s.Start()
		t.Log("Server stopped.")
	}()

	c := accord.NewAccordClient(serverID)
	c.Connect(serverAddr)

	err = c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw1")
	require.NoError(t, err)
}

func TestClientLogin(t *testing.T) {
	t.Parallel()

	serverID := uint64(12345)
	s := accord.NewAccordServer()
	serverAddr, err := s.Listen("localhost:0")
	go func() {
		s.Start()
		t.Log("Server stopped.")
	}()

	c := accord.NewAccordClient(serverID)
	c.Connect(serverAddr)

	err = c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw2")
	require.NotNil(t, err)

	err = c.Login("testuser2", "testpw1")
	require.NotNil(t, err)

	err = c.Login("testuser1", "testpw1")
	require.Nil(t, err)
}

func TestClientCreateChannel(t *testing.T) {
	t.Parallel()

	serverID := uint64(12345)
	s := accord.NewAccordServer()
	serverAddr, err := s.Listen("localhost:0")
	go func() {
		s.Start()
		t.Log("Server stopped.")
	}()

	c := accord.NewAccordClient(serverID)
	c.Connect(serverAddr)

	err = c.CreateChannel("testchan1", true)
	require.NotNil(t, err)

	err = c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.CreateChannel("testchan1", true)
	require.NoError(t, err)
}
