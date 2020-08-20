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

// Channel represents a single private or public messaging channel.
type Channel struct {
	channelID           uint64
	name                string
	messages            []Message
	msgc                chan Message
	usersToStreams      map[string]pb.Chat_StreamServer
	users               []User
	pinnedMsg           uint64
	isPublic            bool
	rolesWithPermission map[Permission][]string
}

// NewChannel creates a new channel with provided parameters.
func NewChannel(uid uint64, name string, isPublic bool) *Channel {
	return &Channel{
		channelID:           uid,
		name:                name,
		messages:            []Message{},
		msgc:                make(chan Message),
		usersToStreams:      map[string]pb.Chat_StreamServer{},
		pinnedMsg:           0,
		isPublic:            isPublic,
		rolesWithPermission: map[Permission][]string{},
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
		username := "tmr"
		if stream := ch.usersToStreams[username]; stream != nil {
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
				log.Printf("Could not send message to %s in channel %v\n", user.username, *ch)
			}
		}
	}
}
