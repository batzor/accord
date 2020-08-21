package accord

import (
	"log"

	pb "github.com/qvntm/accord/pb"
)

// Channel represents a single private or public messaging channel.
type Channel struct {
	channelID           uint64
	name                string
	msgc                chan RequestMessage
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
		msgc:                make(chan RequestMessage),
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
		case <-ch.msgc:
			// TODO: update this when I think of message type.
			ch.Broadcast(nil)
		}
	}
}

// Broadcast sends message to all users in the chat.
func (ch *Channel) Broadcast(msg *ResponseMessage) {
	for _, user := range ch.users {
		// TODO: also check for permissions to read (i.e. receive broadcast)
		// Checking whether user is subscribed to the channel at the moment
		if stream := ch.usersToStreams[user.username]; stream != nil {
			response, _ := msg.getStreamResponse()
			if err := stream.Send(response); err != nil {
				log.Printf("Could not send message to %s in channel %v\n", user.username, *ch)
			}
		}
	}
}
