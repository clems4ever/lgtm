package protocol

import "github.com/clems4ever/lgtm/internal/github"

// ApproveRequestMessage is sent to request or notify about a PR approval.
type ApproveRequestMessage struct {
	// Link contains the information about the pull request to be approved.
	Link github.PRLink
}

// ApproveResponseType represents the type of response to an approval request.
type ApproveResponseType string

const (
	// ApproveResponseErrSameAuthor indicates the approver is the same as the PR author (not allowed).
	ApproveResponseErrSameAuthor ApproveResponseType = "error_same_author"
	// ApproveResponseSuccess indicates the PR was successfully approved.
	ApproveResponseSuccess ApproveResponseType = "success"
)

// ApproveResponseMessage is sent in response to an ApproveRequestMessage.
type ApproveResponseMessage struct {
	// Response indicates the result of the approval attempt.
	Response ApproveResponseType `json:"response"`
}
