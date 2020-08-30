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

// TestClientCreateChannel checks that the client can create and remove
// one channel. It also ensures that channel creation will fail before
// client has logged in.
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

	// Channel creation has to fail when user is
	// not logged in.
	_, err = c.CreateChannel("testchan1", true)
	require.NotNil(t, err)

	err = c.CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	err = c.Login("testuser1", "testpw1")
	require.NoError(t, err)

	channelID, err := c.CreateChannel("testchan1", true)
	require.NoError(t, err)

	err = c.RemoveChannel(channelID)
	require.NoError(t, err)
}

// TestClientCreateManyChannels checks for creation and removal
// of multiple channels by a single client.
func TestClientCreateManyChannels(t *testing.T) {
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

	id1, err := c.CreateChannel("testChan1", true)
	require.NoError(t, err)

	err = c.RemoveChannel(id1)
	require.NoError(t, err)

	id2, err := c.CreateChannel("testChan2", false)
	require.NoError(t, err)

	id3, err := c.CreateChannel("testChan3", true)
	require.NoError(t, err)

	err = c.RemoveChannel(id2)
	require.NoError(t, err)

	err = c.RemoveChannel(id3)
	require.NoError(t, err)
}
