package client

import "github.com/qvntm/Accord/pb"

type MessageType int

const (
	FromUserMessageType MessageType = iota
	ConfMessageType
)

type Message struct {
	Msg isMessage_Msg
}

type isMessage_Msg interface {
	isMessage_Msg()
}

type UserMessageType int

const (
	SendUserMessageType UserMessageType = iota
	EditUserMessageType
	DeleteUserMessageType
)

var UserToPBMessages = map[UserMessageType]pb.StreamRequest_UserMsgType{
	SendUserMessageType:   pb.StreamRequest_SEND_MSG,
	EditUserMessageType:   pb.StreamRequest_EDIT_MSG,
	DeleteUserMessageType: pb.StreamRequest_DELETE_MSG,
}

type UserMessage struct {
	Type    UserMessageType
	Content string
}

func (*UserMessage) isMessage_Msg() {}

func (m *Message) GetMsg() isMessage_Msg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (x *Message) GetUserMsg() *UserMessage {
	if x, ok := x.GetMsg().(*UserMessage); ok {
		return x
	}
	return nil
}
