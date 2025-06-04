package protocol

// MessageType represents the type of a protocol message.
type MessageType string

const (
	// ApproveMessageType is sent to request or notify about a PR approval.
	ApproveRequestMessageType  MessageType = "approve_request"
	ApproveResponseMessageType MessageType = "approve_response"
	// RegisterRequestMessageType is sent by a client to register itself as an approver.
	RegisterRequestMessageType MessageType = "register_request"
	PingMessageType            MessageType = "ping"
)

// Message is a generic wrapper for protocol messages exchanged over the websocket.
type Message struct {
	Type      MessageType // The type of the message (e.g., "approve", "register").
	RequestID string      // A request ID to link the potential responses.
	Message   any         // The actual message payload (e.g., ApproveMessage, RegisterMessage).
}
