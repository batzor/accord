package tests

import (
	"testing"

	client "github.com/qvntm/Accord/client"
	server "github.com/qvntm/Accord/server"
	"github.com/stretchr/testify/require"
)

func TestClientCreateUser(t *testing.T) {
	t.Parallel()

	s := server.NewAccordServer()
	addr, err := s.Start("0.0.0.0:0")
	require.NoError(t, err)

	c := client.NewAccordClient()
	c.Connect(addr)

	err = c.AuthClient().CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	_, err = c.AuthClient().Login("testuser1", "testpw1")
	require.NoError(t, err)
}

func TestClientLogin(t *testing.T) {
	t.Parallel()

	s := server.NewAccordServer()
	addr, err := s.Start("0.0.0.0:0")

	c := client.NewAccordClient()
	c.Connect(addr)

	err = c.AuthClient().CreateUser("testuser1", "testpw1")
	require.NoError(t, err)

	_, err = c.AuthClient().Login("testuser1", "testpw2")
	require.NotNil(t, err)

	_, err = c.AuthClient().Login("testuser2", "testpw1")
	require.NotNil(t, err)

	_, err = c.AuthClient().Login("testuser1", "testpw1")
	require.Nil(t, err)
}
