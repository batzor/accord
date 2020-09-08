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

	username := accord.GetRandUsername()
	password := accord.GetRandPassword()
	err = c.CreateUser(username, password)
	require.NoError(t, err)

	err = c.Login(username, password)
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

	username := accord.GetRandUsername()
	password := accord.GetRandPassword()
	channelName := accord.GetRandChannelName()
	isPublic := accord.GetRandBool()
	// Channel creation has to fail when user is
	// not logged in.
	_, err = c.CreateChannel(channelName, isPublic)
	require.NotNil(t, err)

	err = c.CreateUser(username, password)
	require.NoError(t, err)

	err = c.Login(username, password)
	require.NoError(t, err)

	channelID, err := c.CreateChannel(channelName, isPublic)
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

	username := accord.GetRandUsername()
	password := accord.GetRandPassword()

	err = c.CreateUser(username, password)
	require.NoError(t, err)

	err = c.Login(username, password)
	require.NoError(t, err)

	id1, err := c.CreateChannel(accord.GetRandChannelName(), accord.GetRandBool())
	require.NoError(t, err)

	err = c.RemoveChannel(id1)
	require.NoError(t, err)

	id2, err := c.CreateChannel(accord.GetRandChannelName(), accord.GetRandBool())
	require.NoError(t, err)

	id3, err := c.CreateChannel(accord.GetRandChannelName(), accord.GetRandBool())
	require.NoError(t, err)

	err = c.RemoveChannel(id2)
	require.NoError(t, err)

	err = c.RemoveChannel(id3)
	require.NoError(t, err)
}

func TestClientChannelStream(t *testing.T) {
	t.Parallel()

	serverID := uint64(12345)
	s := accord.NewAccordServer()
	serverAddr, err := s.Listen("localhost:0")
	go func() {
		s.Start()
		t.Log("Server stopped.")
	}()

	// create first user and login
	c1 := accord.NewAccordClient(serverID)
	c1.Connect(serverAddr)
	username1 := accord.GetRandUsername()
	password1 := accord.GetRandPassword()
	err = c1.CreateUser(username1, password1)
	require.NoError(t, err)
	err = c1.Login(username1, password1)
	require.NoError(t, err)

	// create second user and login
	c2 := accord.NewAccordClient(serverID)
	c2.Connect(serverAddr)
	username2 := accord.GetRandUsername()
	password2 := accord.GetRandPassword()
	err = c2.CreateUser(username2, password2)
	require.NoError(t, err)
	err = c2.Login(username2, password2)
	require.NoError(t, err)

	// first user will create a public channel
	channelName := accord.GetRandChannelName()
	isPublic := accord.GetRandBool()
	channelID, err := c1.CreateChannel(channelName, isPublic)
	require.NoError(t, err)

	err = c1.RemoveChannel(channelID)
	require.NoError(t, err)
}
