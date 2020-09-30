package accord

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/golang/protobuf/ptypes"
	pb "github.com/qvntm/accord/pb"
)

type channelUser struct {
	user *User
	role Role
}

// ClientChannel represents a single private or public messaging channel.
type ClientChannel struct {
	ChannelID uint64
	Name      string
	// Set if the **fixed** data for this channel has been fetched on the client side.
	// Data that is mutable (and is frequently updated) such as pinned message, channel name,
	// and messages, is not polled through "IsFetched".
	IsFetched           bool
	Users               map[string]*channelUser
	PinnedMsgID         uint64
	IsPublic            bool
	RolesWithPermission map[Permission][]Role
	Messages            []Message
	Stream              pb.Chat_ChannelStreamClient
}

// ServerChannel represents a single private or public messaging channel.
type ServerChannel struct {
	channelID uint64
	name      string
	msgc      chan *pb.ChannelStreamRequest
	// users contains general information about users in the channel
	users map[string]*channelUser
	// usersToStreams has only streams of users, which are streaming at the moment
	usersToStreams      map[string]pb.Chat_ChannelStreamServer
	pinnedMsgID         uint64
	isPublic            bool
	rolesWithPermission map[Permission][]Role
}

// NewClientChannel creates a new client channel with provided parameters.
func NewClientChannel(uid uint64, name string) *ClientChannel {
	return &ClientChannel{
		ChannelID: uid,
		Name:      name,
		IsFetched: false,
	}
}

// NewServerChannel creates a new server channel with provided parameters.
func NewServerChannel(uid uint64, name string, isPublic bool) *ServerChannel {
	return &ServerChannel{
		channelID:           uid,
		name:                name,
		msgc:                make(chan *pb.ChannelStreamRequest),
		users:               make(map[string]*channelUser),
		pinnedMsgID:         0,
		isPublic:            isPublic,
		rolesWithPermission: make(map[Permission][]Role),
	}
}

func (ch *ServerChannel) addUser(user *channelUser) {
	ch.users[user.user.username] = user
}

// Listen listens for the incoming messages.
func (ch *ServerChannel) listen() {
	for {
		select {
		case req := <-ch.msgc:
			res, err := ch.processChannelStreamRequest(req)
			if err == nil {
				ch.broadcast(res)
			} else {
				log.Printf("Failed to process request %v: %v\n", req, err)
			}
		}
	}
}

// Broadcast sends message to all users in the chat.
func (ch *ServerChannel) broadcast(response *pb.ChannelStreamResponse) {
	// only broadcast to clients, who are currently streaming with the server
	for username, stream := range ch.usersToStreams {
		// TODO: also check for permissions to read (i.e. receive broadcast)
		if err := stream.Send(response); err != nil {
			log.Printf("Could not send message to %s in channel %v\n", username, ch.name)
		}
	}
}

func (ch *ServerChannel) processChannelStreamRequest(m *pb.ChannelStreamRequest) (*pb.ChannelStreamResponse, error) {
	switch m.GetMsg().(type) {
	case *pb.ChannelStreamRequest_UserMsg:
		res, err := ch.processChannelStreamRequestUserMessage(m.GetUserMsg())
		if err != nil {
			return &pb.ChannelStreamResponse{
				Msg: &pb.ChannelStreamResponse_UserMsg{
					UserMsg: res,
				},
			}, nil
		}
		return nil, err
	case *pb.ChannelStreamRequest_ConfigMsg:
		res, err := ch.processChannelStreamRequestConfigMessage(m.GetConfigMsg())
		if err != nil {
			return &pb.ChannelStreamResponse{
				Msg: &pb.ChannelStreamResponse_ConfigMsg{
					ConfigMsg: res,
				},
			}, nil
		}
		return nil, err
	}
	return nil, fmt.Errorf("Invalid request type: %v", reflect.TypeOf(m.GetMsg()))
}

// TODO: Totally rewrite this function when we add persistent layer.
func (ch *ServerChannel) processChannelStreamRequestUserMessage(m *pb.ChannelStreamRequest_UserMessage) (*pb.ChannelStreamResponse_UserMessage, error) {
	switch m.GetUserMsg().(type) {
	case *pb.ChannelStreamRequest_UserMessage_NewUserMsg:
		timestamp, _ := ptypes.TimestampProto(time.Now())
		return &pb.ChannelStreamResponse_UserMessage{
			MessageId: rand.Uint64(),
			UserMsg: &pb.ChannelStreamResponse_UserMessage_NewAndUpdateUserMsg{
				NewAndUpdateUserMsg: &pb.ChannelStreamResponse_UserMessage_NewAndUpdateUserMessage{
					Timestamp: timestamp,
					Content:   m.GetNewUserMsg().GetContent(),
				},
			},
		}, nil
	case *pb.ChannelStreamRequest_UserMessage_EditUserMsg:
		return nil, fmt.Errorf("persistent layer is not implemented yet, thus, message editing is not implemented too")
	case *pb.ChannelStreamRequest_UserMessage_DeleteUserMsg:
		return nil, fmt.Errorf("persistent layer is not implemented yet, thus, message deletion is not implemented too")
	}
	return nil, fmt.Errorf("Invalid object type: %v", reflect.TypeOf(m.GetUserMsg()))
}

func (ch *ServerChannel) processChannelStreamRequestConfigMessage(m *pb.ChannelConfigMessage) (*pb.ChannelConfigMessage, error) {
	switch m.GetMsg().(type) {
	case *pb.ChannelConfigMessage_NameMsg:
		nameMsg := m.GetNameMsg()
		ch.name = nameMsg.GetNewChannelName()
		return m, nil
	case *pb.ChannelConfigMessage_RoleMsg:
		roleMsg := m.GetRoleMsg()
		user := ch.users[roleMsg.GetUsername()]
		if user != nil {
			user.role = PBToAccordRoles[roleMsg.GetRole()]
		}
		return nil, fmt.Errorf("user '%s' is not in the channel %s", roleMsg.GetUsername(), ch.name)
	case *pb.ChannelConfigMessage_PinMsg:
		pinMsg := m.GetPinMsg()
		ch.pinnedMsgID = pinMsg.GetMessageId()
		return m, nil
	}
	return nil, fmt.Errorf("Invalid object type: %v", reflect.TypeOf(m.GetMsg()))
}
