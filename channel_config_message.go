package accord

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
