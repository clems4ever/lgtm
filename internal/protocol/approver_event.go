package protocol

// ApproverEventMessage is used to communicate the list of available approvers.
type ApproverEventMessage struct {
	Approvers []string
}
