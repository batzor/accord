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

type ChannelUser struct {
	user *User
	role Role
}

// Channel represents a single private or public messaging channel.
type Channel struct {
	channelID uint64
	name      string
	msgc      chan *pb.ChannelStreamRequest
	// users contains general information about users in the channel
	users map[string]*ChannelUser
	// usersToStreams has only streams of users, which are streaming at the moment
	usersToStreams      map[string]pb.Chat_ChannelStreamServer
	pinnedMsgID         uint64
	isPublic            bool
	rolesWithPermission map[Permission][]Role
}

// NewChannel creates a new channel with provided parameters.
func NewChannel(uid uint64, name string, isPublic bool) *Channel {
	return &Channel{
		channelID:           uid,
		name:                name,
		msgc:                make(chan *pb.ChannelStreamRequest),
		users:               make(map[string]*ChannelUser),
		pinnedMsgID:         0,
		isPublic:            isPublic,
		rolesWithPermission: make(map[Permission][]Role),
	}
}

// Listen listens for the incoming messages.
func (ch *Channel) listen() {
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
func (ch *Channel) broadcast(response *pb.ChannelStreamResponse) {
	// only broadcast to clients, who are currently streaming with the server
	for username, stream := range ch.usersToStreams {
		// TODO: also check for permissions to read (i.e. receive broadcast)
		if err := stream.Send(response); err != nil {
			log.Printf("Could not send message to %s in channel %v\n", username, ch.name)
		}
	}
}

func (ch *Channel) processChannelStreamRequest(m *pb.ChannelStreamRequest) (*pb.ChannelStreamResponse, error) {
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
func (ch *Channel) processChannelStreamRequestUserMessage(m *pb.ChannelStreamRequest_UserMessage) (*pb.ChannelStreamResponse_UserMessage, error) {
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

func (ch *Channel) processChannelStreamRequestConfigMessage(m *pb.ChannelConfigMessage) (*pb.ChannelConfigMessage, error) {
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
