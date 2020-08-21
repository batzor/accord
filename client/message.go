package client

import (
	"fmt"
	"reflect"
	"time"

	"github.com/qvntm/Accord/pb"
)

// RequestMessageType is a type of a stream request message.
type RequestMessageType int

const (
	// FromUserRequestMessageType is message sent by any user to the channel.
	FromUserRequestMessageType RequestMessageType = iota
	// ConfChangeRequestMessageType carries channel/server configuration changes.
	ConfChangeRequestMessageType
)

type RequestMessage struct {
	Username  string
	ChannelID uint64
	Msg       isRequestMessageMsg
}

type isRequestMessageMsg interface {
	isRequestMessageMsg()
}

// UserRequestMessage is a stream message sent by one of the users to the channel.
type UserRequestMessage struct {
	UserMsg isUserRequestMessageUserMsg
}

type isUserRequestMessageUserMsg interface {
	isUserRequestMessageUserMsg()
}

func (*UserRequestMessage) isRequestMessageMsg() {}

type NewUserRequestMessage struct {
	Content string
}

func (*NewUserRequestMessage) isUserRequestMessageUserMsg() {}

func (m *NewUserRequestMessage) getStreamRequestUserMessageNewUserMsg() *pb.StreamRequest_UserMessage_NewUserMsg {
	return &pb.StreamRequest_UserMessage_NewUserMsg{
		NewUserMsg: &pb.StreamRequest_UserMessage_NewUserMessage{
			Content: m.Content,
		},
	}
}

type EditUserRequestMessage struct {
	MessageID uint64
	Content   string
}

func (*EditUserRequestMessage) isUserRequestMessageUserMsg() {}

func (m *EditUserRequestMessage) getStreamRequestUserMessageEditUserMsg() *pb.StreamRequest_UserMessage_EditUserMsg {
	return &pb.StreamRequest_UserMessage_EditUserMsg{
		EditUserMsg: &pb.StreamRequest_UserMessage_EditUserMessage{
			MessageId: m.MessageID,
			Content:   m.Content,
		},
	}
}

type DeleteUserRequestMessage struct {
	MessageID uint64
}

func (*DeleteUserRequestMessage) isUserRequestMessageUserMsg() {}

func (m *DeleteUserRequestMessage) getStreamRequestUserMessageDeleteUserMsg() *pb.StreamRequest_UserMessage_DeleteUserMsg {
	return &pb.StreamRequest_UserMessage_DeleteUserMsg{
		DeleteUserMsg: &pb.StreamRequest_UserMessage_DeleteUserMessage{
			MessageId: m.MessageID,
		},
	}
}

// GetUserMsg returns a user messega carried by m.
func (m *UserRequestMessage) GetUserMsg() isUserRequestMessageUserMsg {
	if m != nil {
		return m.UserMsg
	}
	return nil
}

// getStreamRequestUserMsg turns user request message to the similar message declared
// by pb.go file from "pb" package.
func (m *UserRequestMessage) getStreamRequestUserMsg() *pb.StreamRequest_UserMsg {
	var userMsg *pb.StreamRequest_UserMsg
	switch m.GetUserMsg().(type) {
	case *NewUserRequestMessage:
		userMsg = &pb.StreamRequest_UserMsg{
			UserMsg: &pb.StreamRequest_UserMessage{
				UserMsg: m.GetNewUserMsg().getStreamRequestUserMessageNewUserMsg(),
			},
		}
	case *EditUserRequestMessage:
		userMsg = &pb.StreamRequest_UserMsg{
			UserMsg: &pb.StreamRequest_UserMessage{
				UserMsg: m.GetEditUserMsg().getStreamRequestUserMessageEditUserMsg(),
			},
		}
	case *DeleteUserRequestMessage:
		userMsg = &pb.StreamRequest_UserMsg{
			UserMsg: &pb.StreamRequest_UserMessage{
				UserMsg: m.GetDeleteUserMsg().getStreamRequestUserMessageDeleteUserMsg(),
			},
		}
	}
	return userMsg
}

func (m *UserRequestMessage) GetNewUserMsg() *NewUserRequestMessage {
	if x, ok := m.GetUserMsg().(*NewUserRequestMessage); ok {
		return x
	}
	return nil
}

func (m *UserRequestMessage) GetEditUserMsg() *EditUserRequestMessage {
	if x, ok := m.GetUserMsg().(*EditUserRequestMessage); ok {
		return x
	}
	return nil
}

func (m *UserRequestMessage) GetDeleteUserMsg() *DeleteUserRequestMessage {
	if x, ok := m.GetUserMsg().(*DeleteUserRequestMessage); ok {
		return x
	}
	return nil
}

type ConfRequestMessageType int

const (
	EditChannelConfRequestMessageType ConfRequestMessageType = iota
	EditServerConfRequestMessageType
)

var ConfRequestToPBMessages = map[ConfRequestMessageType]pb.StreamRequest_ConfMsgType{
	EditChannelConfRequestMessageType: pb.StreamRequest_EDIT_CHANNEL,
	EditServerConfRequestMessageType:  pb.StreamRequest_EDIT_SERVER,
}

type ConfRequestMessage struct {
	MsgType     ConfRequestMessageType
	Placeholder string
}

func (*ConfRequestMessage) isRequestMessageMsg() {}

func (m *RequestMessage) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *RequestMessage) GetChannelID() uint64 {
	if m != nil {
		return m.ChannelID
	}
	return 0
}

func (m *RequestMessage) GetMsg() isRequestMessageMsg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (x *RequestMessage) GetUserMsg() *UserRequestMessage {
	if x, ok := x.GetMsg().(*UserRequestMessage); ok {
		return x
	}
	return nil
}

func (x *RequestMessage) GetConfMsg() *ConfRequestMessage {
	if x, ok := x.GetMsg().(*ConfRequestMessage); ok {
		return x
	}
	return nil
}

func (m *RequestMessage) getStreamRequest() (*pb.StreamRequest, error) {
	switch m.GetMsg().(type) {
	case *UserRequestMessage:
		return &pb.StreamRequest{
			Username:  m.Username,
			ChannelId: m.ChannelID,
			Msg:       m.GetUserMsg().getStreamRequestUserMsg(),
		}, nil
	case *ConfRequestMessage:
		return &pb.StreamRequest{
			Username:  m.Username,
			ChannelId: m.ChannelID,
			Msg:       nil, // TODO: implement this.
		}, nil
	}
	return nil, fmt.Errorf("invalid message type was passed: %v", reflect.TypeOf(m.GetMsg()))
}

// ResponseMessageType is a type of a stream response message.
type ResponseMessageType int

const (
	// NewMessageResponseMessageType represents new messages sent to chat.
	NewMessageResponseMessageType ResponseMessageType = iota
	// UpdateMessageResponseMessageType represents changes to existing messages.
	UpdateMessageResponseMessageType
)

type ResponseMessage struct {
	Timestamp time.Time
	Msg       isResponseMessageMsg
}

type isResponseMessageMsg interface {
	isResponseMessageMsg()
}

type NewMessageResponseMessage struct {
	SenderID uint64
	Content  string
}

func (*NewMessageResponseMessage) isResponseMessageMsg() {}

type UpdateMessageResponseMessage struct {
	Placeholder string
}

func (*UpdateMessageResponseMessage) isResponseMessageMsg() {}

func (m *ResponseMessage) GetTimestamp() *time.Time {
	if m != nil {
		return &m.Timestamp
	}
	return nil
}

func (m *ResponseMessage) GetMsg() isResponseMessageMsg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (x *ResponseMessage) GetNewMessageMsg() *NewMessageResponseMessage {
	if x, ok := x.GetMsg().(*NewMessageResponseMessage); ok {
		return x
	}
	return nil
}

func (x *ResponseMessage) GetUpdateMessageMsg() *UpdateMessageResponseMessage {
	if x, ok := x.GetMsg().(*UpdateMessageResponseMessage); ok {
		return x
	}
	return nil
}
