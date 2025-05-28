package protocol

// MessageType represents the type of a protocol message.
type MessageType string

const (
	// ApproveMessageType is sent to request or notify about a PR approval.
	ApproveMessageType MessageType = "approve"
	// RegisterRequestMessageType is sent by a client to register itself as an approver.
	RegisterRequestMessageType MessageType = "register_request"
	// ApproverEventMessageType is sent to communicate the list of available approvers.
	ApproverEventMessageType MessageType = "approver_event"
)

// Message is a generic wrapper for protocol messages exchanged over the websocket.
type Message struct {
	Type    MessageType // The type of the message (e.g., "approve", "register").
	Message any         // The actual message payload (e.g., ApproveMessage, RegisterMessage).
}
