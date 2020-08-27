package accord

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/qvntm/accord/pb"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

type AccordServer struct {
	authServer      *AuthServer
	authInterceptor *ServerAuthInterceptor
	listener        net.Listener
	mutex           sync.RWMutex
	channels        map[uint64]*Channel
	jwtManager      *JWTManager
}

func NewAccordServer() *AccordServer {
	authServer := NewAuthServer()
	return &AccordServer{
		authServer:      authServer,
		authInterceptor: NewServerAuthInterceptor(authServer.JWTManager()),
		channels:        make(map[uint64]*Channel),
		jwtManager:      NewJWTManager(secretKey, tokenDuration),
	}
}

// LoadChannels loads channels from the persistent storage
func (s *AccordServer) LoadChannels() error {
	return fmt.Errorf("unimplemented")
}

// LoadUsers loads channels from the persistent storage
func (s *AccordServer) LoadUsers() error {
	return fmt.Errorf("unimplemented")
}

func (s *AccordServer) AddChannel(_ context.Context, req *pb.AddChannelRequest) (*pb.AddChannelResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch := NewChannel(uint64(len(s.channels)), req.GetName(), req.GetIsPublic())
	// TODO: add the new channel to the DB.
	// TODO: broadcast to ServerStream creation of new channel.
	s.channels[ch.channelID] = ch
	go ch.listen()

	res := &pb.AddChannelResponse{}
	log.Printf("New Channel %s created", req.GetName())
	return res, nil
}

func (s *AccordServer) RemoveChannel(_ context.Context, req *pb.RemoveChannelRequest) (*pb.RemoveChannelResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	channelID := req.GetChannelId()
	if _, ok := s.channels[req.GetChannelId()]; ok {
		// TODO: remove the record from the DB.
		// TODO: broadcast to ServerStream removal of the channel.
		delete(s.channels, req.GetChannelId())
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "channel with ID %d doesn't exist", channelID)
	}

	res := &pb.RemoveChannelResponse{}
	log.Printf("Channel with id %d has been removed\n", channelID)
	return res, nil
}

// ChannelStream is the implementation of bidirectional streaming of client
// with one channel on the server.
func (s *AccordServer) ChannelStream(srv pb.Chat_ChannelStreamServer) error {
	var channel *Channel = nil
	var username string = ""
	ctx := srv.Context()

	for {
		req, err := srv.Recv()
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		if username == "" {
			username = req.GetUsername()
		} else if reqUsername := req.GetUsername(); username != reqUsername {
			return status.Errorf(codes.InvalidArgument, "each stream has to use consistent usernames\nhave:%s\nwant:%s\n", reqUsername, username)
		}

		if channel == nil {
			channel = s.channels[req.GetChannelId()]
			if channel == nil {
				return status.Errorf(codes.InvalidArgument, "invalid channel ID: %v", err)
			}
			channel.usersToStreams[username] = srv
			// so far, authomatically add user as a member when he subscribes to the channel
			// TODO: add some RPC for user to request to join the channel with particular role.
			channel.users[username] = &ChannelUser{
				user: s.authServer.users[username],
				role: MemberRole,
			}
		} else if reqChannelID := req.GetChannelId(); channel.channelID != reqChannelID {
			return status.Errorf(codes.InvalidArgument, "each stream has to use consistent channel IDs\nhave:%d\nwant:%d\n", reqChannelID, channel.channelID)
		}

		select {
		// handle abrupt client disconnection
		case <-ctx.Done():
			channel.usersToStreams[username] = nil
			return status.Error(codes.Canceled, ctx.Err().Error())
		case channel.msgc <- req:
		}
	}

	return nil
}

func (s *AccordServer) Listen(serv_addr string) (string, error) {
	listener, err := net.Listen("tcp", serv_addr)
	if err != nil {
		log.Print("Failed to init listener:", err)
		return "", err
	}
	log.Print("Initialized listener:", listener.Addr().String())

	s.listener = listener
	return s.listener.Addr().String(), nil
}

func (s *AccordServer) Start() {
	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("Cannot load TLS credentials:", err)
	}
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(s.authInterceptor.Unary()),
		grpc.StreamInterceptor(s.authInterceptor.Stream()),
	}

	serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))

	srv := grpc.NewServer(serverOptions...)
	pb.RegisterAuthServiceServer(srv, s.authServer)
	pb.RegisterChatServer(srv, s)

	srv.Serve(s.listener)
}
