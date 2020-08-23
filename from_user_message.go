package accord

import (
	"time"
)

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
