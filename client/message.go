package client

import (
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

type UserRequestMessageType int

const (
	SendUserRequestMessageType UserRequestMessageType = iota
	EditUserRequestMessageType
	DeleteUserRequestMessageType
)

var UserRequestToPBMessages = map[UserRequestMessageType]pb.StreamRequest_UserMsgType{
	SendUserRequestMessageType:   pb.StreamRequest_SEND_MSG,
	EditUserRequestMessageType:   pb.StreamRequest_EDIT_MSG,
	DeleteUserRequestMessageType: pb.StreamRequest_DELETE_MSG,
}

type UserRequestMessage struct {
	MsgType UserRequestMessageType
	Content string
}

func (*UserRequestMessage) isRequestMessageMsg() {}

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
