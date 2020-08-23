package accord

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
