package server

// Message represents a single message in a channel.
type Message struct {
	timestamp uint64
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
	return
}
