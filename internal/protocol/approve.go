package protocol

import "github.com/clems4ever/lgtm/internal/github"

// ApproveMessage is sent to request or notify about a PR approval.
type ApproveMessage struct {
	// Link to the PR.
	Link github.PRLink
	// The author of the PR. It's a github username.
	Author string
}
