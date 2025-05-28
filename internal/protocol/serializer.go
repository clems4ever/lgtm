package protocol

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func Read(conn *websocket.Conn, msg *Message) error {
	err := conn.ReadJSON(msg)
	if err != nil {
		return fmt.Errorf("failed to read json message: %w", err)
	}

	mb, err := json.Marshal(msg.Message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	switch msg.Type {
	case ApproveMessageType:
		var approvalMessage ApproveMessage
		err = json.Unmarshal(mb, &approvalMessage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		msg.Message = approvalMessage
	case RegisterRequestMessageType:
		var tMsg RegisterRequestMessage
		err = json.Unmarshal(mb, &tMsg)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		msg.Message = tMsg
	case ApproverEventMessageType:
		var tMsg ApproverEventMessage
		err = json.Unmarshal(mb, &tMsg)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		msg.Message = tMsg
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
	return nil
}

func Write(conn *websocket.Conn, msg any) error {
	switch v := msg.(type) {
	case ApproveMessage:
		conn.WriteJSON(Message{
			Type:    ApproveMessageType,
			Message: v,
		})
	case RegisterRequestMessage:
		conn.WriteJSON(Message{
			Type:    RegisterRequestMessageType,
			Message: v,
		})
	case ApproverEventMessage:
		conn.WriteJSON(Message{
			Type:    ApproverEventMessageType,
			Message: v,
		})
	default:
		return fmt.Errorf("unsupported msg type")
	}
	return nil

}
