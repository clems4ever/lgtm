package protocol

// RegisterRequestMessage is sent by a client to register itself as an approver for a set of repositories.
// It includes the list of repositories and the GitHub username of the client.
//
// Fields:
// - Repos: A list of repository names in the format "owner/repo".
// - GithubUser: The GitHub username of the client.
type RegisterRequestMessage struct {
	Repos      []string `json:"repos"`
	GithubUser string   `json:"github_user"`
}
