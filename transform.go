package accord

import (
	"github.com/qvntm/accord/pb"
)

func getChannelStreamRequestUserMessageNewUserMsg(m *NewMessageUserChannelStreamRequest) *pb.ChannelStreamRequest_UserMessage_NewUserMsg {
	return &pb.ChannelStreamRequest_UserMessage_NewUserMsg{
		NewUserMsg: &pb.ChannelStreamRequest_UserMessage_NewUserMessage{
			Content: m.Content,
		},
	}
}

func getChannelStreamRequestUserMessageEditUserMsg(m *EditMessageUserChannelStreamRequest) *pb.ChannelStreamRequest_UserMessage_EditUserMsg {
	return &pb.ChannelStreamRequest_UserMessage_EditUserMsg{
		EditUserMsg: &pb.ChannelStreamRequest_UserMessage_EditUserMessage{
			MessageId: m.MessageID,
			Content:   m.Content,
		},
	}
}

func getChannelStreamRequestUserMessageDeleteUserMsg(m *DeleteMessageUserChannelStreamRequest) *pb.ChannelStreamRequest_UserMessage_DeleteUserMsg {
	return &pb.ChannelStreamRequest_UserMessage_DeleteUserMsg{
		DeleteUserMsg: &pb.ChannelStreamRequest_UserMessage_DeleteUserMessage{
			MessageId: m.MessageID,
		},
	}
}

// getChannelStreamRequestUserMsg turns user request message to the similar message declared
// by pb.go file from "pb" package.
func getChannelStreamRequestUserMsg(m *UserChannelStreamRequest) *pb.ChannelStreamRequest_UserMsg {
	var userMsg *pb.ChannelStreamRequest_UserMsg
	switch m.getUserMsg().(type) {
	case *NewMessageUserChannelStreamRequest:
		userMsg = &pb.ChannelStreamRequest_UserMsg{
			UserMsg: &pb.ChannelStreamRequest_UserMessage{
				UserMsg: getChannelStreamRequestUserMessageNewUserMsg(m.getNewUserMsg()),
			},
		}
	case *EditMessageUserChannelStreamRequest:
		userMsg = &pb.ChannelStreamRequest_UserMsg{
			UserMsg: &pb.ChannelStreamRequest_UserMessage{
				UserMsg: getChannelStreamRequestUserMessageEditUserMsg(m.getEditUserMsg()),
			},
		}
	case *DeleteMessageUserChannelStreamRequest:
		userMsg = &pb.ChannelStreamRequest_UserMsg{
			UserMsg: &pb.ChannelStreamRequest_UserMessage{
				UserMsg: getChannelStreamRequestUserMessageDeleteUserMsg(m.getDeleteUserMsg()),
			},
		}
	}
	return userMsg
}

func getChannelConfigMessageNameMsg(m *NameChannelConfigMessage) *pb.ChannelConfigMessage_NameMsg {
	return &pb.ChannelConfigMessage_NameMsg{
		NameMsg: &pb.ChannelConfigMessage_NameChannelConfigMessage{
			NewChannelName: m.NewChannelName,
		},
	}
}

func getChannelConfigMessageRoleMsg(m *RoleChannelConfigMessage) *pb.ChannelConfigMessage_RoleMsg {
	return &pb.ChannelConfigMessage_RoleMsg{
		RoleMsg: &pb.ChannelConfigMessage_RoleChannelConfigMessage{
			Username: m.Username,
			Role:     AccordToPBRoles[m.Role],
		},
	}
}

func getChannelConfigMessagePinMsg(m *PinChannelConfigMessage) *pb.ChannelConfigMessage_PinMsg {
	return &pb.ChannelConfigMessage_PinMsg{
		PinMsg: &pb.ChannelConfigMessage_PinChannelConfigMessage{
			MessageId: m.MessageID,
		},
	}
}

func getChannelStreamRequestConfigMsg(m *ChannelConfigMessage) *pb.ChannelStreamRequest_ConfigMsg {
	switch m.getMsg().(type) {
	case *NameChannelConfigMessage:
		return &pb.ChannelStreamRequest_ConfigMsg{
			ConfigMsg: &pb.ChannelConfigMessage{
				Msg: getChannelConfigMessageNameMsg(m.getNameMsg()),
			},
		}
	case *RoleChannelConfigMessage:
		return &pb.ChannelStreamRequest_ConfigMsg{
			ConfigMsg: &pb.ChannelConfigMessage{
				Msg: getChannelConfigMessageRoleMsg(m.getRoleMsg()),
			},
		}
	case *PinChannelConfigMessage:
		return &pb.ChannelStreamRequest_ConfigMsg{
			ConfigMsg: &pb.ChannelConfigMessage{
				Msg: getChannelConfigMessagePinMsg(m.getPinMsg()),
			},
		}
	}
	return nil
}

func getChannelStreamRequest(m *ChannelStreamRequest) *pb.ChannelStreamRequest {
	switch m.GetMsg().(type) {
	case *UserChannelStreamRequest:
		return &pb.ChannelStreamRequest{
			ChannelId: m.ChannelID,
			Msg:       getChannelStreamRequestUserMsg(m.GetUserMsg()),
		}
	case *ChannelConfigMessage:
		return &pb.ChannelStreamRequest{
			ChannelId: m.ChannelID,
			Msg:       getChannelStreamRequestConfigMsg(m.GetConfMsg()),
		}
	}
	return nil
}

func getNewAndUpdateMessageUserChannelStreamResponse(m *pb.ChannelStreamResponse_UserMessage_NewAndUpdateUserMessage) *NewAndUpdateMessageUserChannelStreamResponse {
	return &NewAndUpdateMessageUserChannelStreamResponse{
		Timestamp: m.GetTimestamp().AsTime(),
		Content:   m.GetContent(),
	}
}

func getDeleteMessageUserChannelStreamResponse(m *pb.ChannelStreamResponse_UserMessage_DeleteUserMessage) *DeleteMessageUserChannelStreamResponse {
	return &DeleteMessageUserChannelStreamResponse{}
}

func getUserChannelStreamResponse(m *pb.ChannelStreamResponse_UserMessage) *UserChannelStreamResponse {
	switch m.GetUserMsg().(type) {
	case *pb.ChannelStreamResponse_UserMessage_NewAndUpdateUserMsg:
		return &UserChannelStreamResponse{
			MessageID: m.GetMessageId(),
			UserMsg:   getNewAndUpdateMessageUserChannelStreamResponse(m.GetNewAndUpdateUserMsg()),
		}
	case *pb.ChannelStreamResponse_UserMessage_DeleteUserMsg:
		return &UserChannelStreamResponse{
			MessageID: m.GetMessageId(),
			UserMsg:   getDeleteMessageUserChannelStreamResponse(m.GetDeleteUserMsg()),
		}
	}
	return nil
}

func getNameChannelConfigMessage(m *pb.ChannelConfigMessage_NameChannelConfigMessage) *NameChannelConfigMessage {
	return &NameChannelConfigMessage{
		NewChannelName: m.GetNewChannelName(),
	}
}

func getRoleChannelConfigMessage(m *pb.ChannelConfigMessage_RoleChannelConfigMessage) *RoleChannelConfigMessage {
	return &RoleChannelConfigMessage{
		Username: m.GetUsername(),
		Role:     PBToAccordRoles[m.GetRole()],
	}
}

func getPinChannelConfigMessage(m *pb.ChannelConfigMessage_PinChannelConfigMessage) *PinChannelConfigMessage {
	return &PinChannelConfigMessage{
		MessageID: m.GetMessageId(),
	}
}

func getChannelConfigMessage(m *pb.ChannelConfigMessage) *ChannelConfigMessage {
	switch m.GetMsg().(type) {
	case *pb.ChannelConfigMessage_NameMsg:
		return &ChannelConfigMessage{
			Msg: getNameChannelConfigMessage(m.GetNameMsg()),
		}
	case *pb.ChannelConfigMessage_RoleMsg:
		return &ChannelConfigMessage{
			Msg: getRoleChannelConfigMessage(m.GetRoleMsg()),
		}
	case *pb.ChannelConfigMessage_PinMsg:
		return &ChannelConfigMessage{
			Msg: getPinChannelConfigMessage(m.GetPinMsg()),
		}
	}
	return nil
}

func getChannelStreamResponse(m *pb.ChannelStreamResponse) *ChannelStreamResponse {
	switch m.GetMsg().(type) {
	case *pb.ChannelStreamResponse_UserMsg:
		return &ChannelStreamResponse{
			Msg: getUserChannelStreamResponse(m.GetUserMsg()),
		}
	case *pb.ChannelStreamResponse_ConfigMsg:
		return &ChannelStreamResponse{
			Msg: getChannelConfigMessage(m.GetConfigMsg()),
		}
	}
	return nil
}
