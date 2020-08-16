package server

import (
	"log"
	"time"

	pb "github.com/qvntm/Accord/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Message represents a single message in a channel.
type Message struct {
	timestamp time.Time
	from      string
	content   string
}

// Permission type represents what user is permitted to do in the channel.
type Permission string

// WritePermission is a permission to write messages to the channel.
var WritePermission Permission = "write_permission"

// ReadPermission is a permission to read messages in the channel.
var ReadPermission Permission = "read_permission"

// Channel represents a single private or public messaging channel.
type Channel struct {
	channelID           uint64
	messages            []Message
	msgc                chan Message
	usersToStreams      map[uint64]pb.Chat_StreamServer
	users               []User
	pinnedMsg           uint64
	isPublic            bool
	rolesWithPermission map[string][]string
}

// NewChannel creates a new channel with provided parameters.
func NewChannel(uid uint64, users []User, isPublic bool) *Channel {
	return &Channel{
		channelID:           uid,
		messages:            []Message{},
		msgc:                make(chan Message),
		usersToStreams:      map[uint64]pb.Chat_StreamServer{},
		users:               users,
		pinnedMsg:           0,
		isPublic:            isPublic,
		rolesWithPermission: map[string][]string{},
	}
}

// Listen listens for the incoming messages.
func (ch *Channel) Listen() {
	for {
		select {
		case msg := <-ch.msgc:
			ch.messages = append(ch.messages, msg)
			ch.Broadcast(msg)
		}
	}
}

// Broadcast sends message to all users in the chat.
func (ch *Channel) Broadcast(msg Message) {
	for _, user := range ch.users {
		// TODO: also check for permissions to read (i.e. receive broadcast)
		userID := uint64(12345)
		if stream := ch.usersToStreams[userID]; stream != nil {
			newMessage := &pb.StreamResponse_NewMsg{
				NewMsg: &pb.StreamResponse_NewMessage{
					SenderId: 12345,
					Content:  msg.content,
				},
			}
			response := &pb.StreamResponse{
				Timestamp: timestamppb.New(msg.timestamp),
				Event:     newMessage,
			}
			if err := stream.Send(response); err != nil {
				log.Printf("Could not send message to %s in channel %v\n", user.Username, *ch)
			}
		}
	}

}
