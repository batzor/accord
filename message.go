package accord

import (
	"time"
)

// ChannelConfigMessage is used in ChannelStreamRequest- and Response
// to initiate and broadcast channel-related changes.
type ChannelConfigMessage struct {
	Msg isChannelConfigMessageMsg
}

type isChannelConfigMessageMsg interface {
	isChannelConfigMessageMsg()
}

func (*ChannelConfigMessage) isChannelStreamRequestMsg() {}

func (*ChannelConfigMessage) isChannelStreamResponseMsg() {}

func (m *ChannelConfigMessage) getMsg() isChannelConfigMessageMsg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (m *ChannelConfigMessage) getNameMsg() *NameChannelConfigMessage {
	if x, ok := m.getMsg().(*NameChannelConfigMessage); ok {
		return x
	}
	return nil
}

func (m *ChannelConfigMessage) getRoleMsg() *RoleChannelConfigMessage {
	if x, ok := m.getMsg().(*RoleChannelConfigMessage); ok {
		return x
	}
	return nil
}

func (m *ChannelConfigMessage) getPinMsg() *PinChannelConfigMessage {
	if x, ok := m.getMsg().(*PinChannelConfigMessage); ok {
		return x
	}
	return nil
}

type NameChannelConfigMessage struct {
	NewChannelName string
}

func (*NameChannelConfigMessage) isChannelConfigMessageMsg() {}

type RoleChannelConfigMessage struct {
	UserID uint64
	Role   Role
}

func (*RoleChannelConfigMessage) isChannelConfigMessageMsg() {}

type PinChannelConfigMessage struct {
	MessageID uint64
}

func (*PinChannelConfigMessage) isChannelConfigMessageMsg() {}

// ChannelStreamRequestType is a type of channel stream request message.
type ChannelStreamRequestType int

const (
	// FromUserChannelStreamRequestType is for sending new messages, and updating or
	// deletion of messages.
	FromUserChannelStreamRequestType ChannelStreamRequestType = iota
	// ChannelConfigChannelStreamRequestType carries channel configuration changes,
	// including name change, updating user roles, and specifying pinned message.
	ChannelConfigChannelStreamRequestType
)

// ChannelStreamRequest represents a stream request for a single channel.
type ChannelStreamRequest struct {
	Username  string
	ChannelID uint64
	Msg       isChannelStreamRequestMsg
}

type isChannelStreamRequestMsg interface {
	isChannelStreamRequestMsg()
}

func (m *ChannelStreamRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *ChannelStreamRequest) GetChannelID() uint64 {
	if m != nil {
		return m.ChannelID
	}
	return 0
}

func (m *ChannelStreamRequest) GetMsg() isChannelStreamRequestMsg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (x *ChannelStreamRequest) GetUserMsg() *UserChannelStreamRequest {
	if x, ok := x.GetMsg().(*UserChannelStreamRequest); ok {
		return x
	}
	return nil
}

func (x *ChannelStreamRequest) GetConfMsg() *ChannelConfigMessage {
	if x, ok := x.GetMsg().(*ChannelConfigMessage); ok {
		return x
	}
	return nil
}

// ChannelStreamResponseType is a type of channel stream response message.
type ChannelStreamResponseType int

const (
	// FromUserChannelStreamResponseType is for sending new messages, and updating or
	// deletion of messages.
	FromUserChannelStreamResponseType ChannelStreamResponseType = iota
	// ChannelConfigChannelStreamResponseType carries channel configuration changes,
	// including name change, updating user roles, and specifying pinned message.
	ChannelConfigChannelStreamResponseType
)

type ChannelStreamResponse struct {
	Msg isChannelStreamResponseMsg
}

type isChannelStreamResponseMsg interface {
	isChannelStreamResponseMsg()
}

// UserChannelStreamRequest is a stream message sent by one of the users to the channel.
type UserChannelStreamRequest struct {
	UserMsg isUserChannelStreamRequestUserMsg
}

type isUserChannelStreamRequestUserMsg interface {
	isUserChannelStreamRequestUserMsg()
}

func (*UserChannelStreamRequest) isChannelStreamRequestMsg() {}

type NewMessageUserChannelStreamRequest struct {
	Content string
}

func (*NewMessageUserChannelStreamRequest) isUserChannelStreamRequestUserMsg() {}

type EditMessageUserChannelStreamRequest struct {
	MessageID uint64
	Content   string
}

func (*EditMessageUserChannelStreamRequest) isUserChannelStreamRequestUserMsg() {}

type DeleteMessageUserChannelStreamRequest struct {
	MessageID uint64
}

func (*DeleteMessageUserChannelStreamRequest) isUserChannelStreamRequestUserMsg() {}

func (m *UserChannelStreamRequest) getUserMsg() isUserChannelStreamRequestUserMsg {
	if m != nil {
		return m.UserMsg
	}
	return nil
}

func (m *UserChannelStreamRequest) getNewUserMsg() *NewMessageUserChannelStreamRequest {
	if x, ok := m.getUserMsg().(*NewMessageUserChannelStreamRequest); ok {
		return x
	}
	return nil
}

func (m *UserChannelStreamRequest) getEditUserMsg() *EditMessageUserChannelStreamRequest {
	if x, ok := m.getUserMsg().(*EditMessageUserChannelStreamRequest); ok {
		return x
	}
	return nil
}

func (m *UserChannelStreamRequest) getDeleteUserMsg() *DeleteMessageUserChannelStreamRequest {
	if x, ok := m.getUserMsg().(*DeleteMessageUserChannelStreamRequest); ok {
		return x
	}
	return nil
}

// UserChannelStreamResponse is a stream message broadcasted to all users in the channel.
type UserChannelStreamResponse struct {
	MessageID uint64
	UserMsg   isUserChannelStreamResponseUserMsg
}

type isUserChannelStreamResponseUserMsg interface {
	isUserChannelStreamResponseUserMsg()
}

func (*UserChannelStreamResponse) isChannelStreamResponseMsg() {}

type NewAndUpdateMessageUserChannelStreamResponse struct {
	Timestamp time.Time
	Content   string
}

func (*NewAndUpdateMessageUserChannelStreamResponse) isUserChannelStreamResponseUserMsg() {}

type DeleteMessageUserChannelStreamResponse struct{}

func (*DeleteMessageUserChannelStreamResponse) isUserChannelStreamResponseUserMsg() {}

// GetMessageID returns message ID or the user stream response. It returns 0
// if the user stream response is nil.
func (m *UserChannelStreamResponse) GetMessageID() uint64 {
	if m != nil {
		return m.MessageID
	}
	return 0
}

// GetUserMsg returns the user message contained in the user channel stream response.
func (m *UserChannelStreamResponse) GetUserMsg() isUserChannelStreamResponseUserMsg {
	if m != nil {
		return m.UserMsg
	}
	return nil
}

// GetNewAndUpdateUserMsg gets the new and update response messages from some user.
func (m *UserChannelStreamResponse) GetNewAndUpdateUserMsg() *NewAndUpdateMessageUserChannelStreamResponse {
	if x, ok := m.GetUserMsg().(*NewAndUpdateMessageUserChannelStreamResponse); ok {
		return x
	}
	return nil
}

// GetDeleteUserMsg gets the deleted message in user response message.
func (m *UserChannelStreamResponse) GetDeleteUserMsg() *DeleteMessageUserChannelStreamResponse {
	if x, ok := m.GetUserMsg().(*DeleteMessageUserChannelStreamResponse); ok {
		return x
	}
	return nil
}
