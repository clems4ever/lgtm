package protocol

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
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
	case ApproveRequestMessageType:
		var approvalMessage ApproveRequestMessage
		err = json.Unmarshal(mb, &approvalMessage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		msg.Message = approvalMessage
	case ApproveResponseMessageType:
		var approvalMessage ApproveResponseMessage
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
	case PingMessageType:
		var tMsg PingMessage
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

// Write sends a message with a generated request ID and returns the request ID.
func Write(conn *websocket.Conn, msg any) (string, error) {
	requestID := uuid.NewString()
	err := WriteWithRequestID(conn, msg, requestID)
	return requestID, err
}

// WriteWithRequestID sends a message with a provided request ID.
func WriteWithRequestID(conn *websocket.Conn, msg any, requestID string) error {
	switch v := msg.(type) {
	case ApproveRequestMessage:
		return conn.WriteJSON(Message{
			Type:      ApproveRequestMessageType,
			RequestID: requestID,
			Message:   v,
		})
	case ApproveResponseMessage:
		return conn.WriteJSON(Message{
			Type:      ApproveResponseMessageType,
			RequestID: requestID,
			Message:   v,
		})
	case RegisterRequestMessage:
		return conn.WriteJSON(Message{
			Type:      RegisterRequestMessageType,
			RequestID: requestID,
			Message:   v,
		})
	case PingMessage:
		return conn.WriteJSON(Message{
			Type:      PingMessageType,
			RequestID: requestID,
			Message:   v,
		})
	default:
		return fmt.Errorf("unsupported msg type")
	}
}
