package protocol

// PingMessage is a protocol message used to keep the WebSocket connection alive.
//
// This message can be sent periodically by either the client or the server to prevent
// idle timeouts and to detect broken connections. It does not carry any payload and
// should be handled as a no-op by the receiver.
type PingMessage struct{}
