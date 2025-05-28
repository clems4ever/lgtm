package client

import (
	"log"
	"net/http"

	"github.com/clems4ever/lgtm/internal/github"
	"github.com/clems4ever/lgtm/internal/protocol"
)

// HandlerApprove handles POST requests to submit a PR for approval.
// It parses the PR link, retrieves the PR author, and sends an ApproveMessage over the websocket.
func (c *Client) handlerApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	prLink := r.FormValue("pr_link")
	link, err := github.ParsePullRequestURL(prLink)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid PR link", http.StatusBadRequest)
		return
	}

	// Retrieve the PR author's GitHub username.
	author, err := c.githubClient.GetPRAuthor(link)
	if err != nil {
		log.Printf("failed to get PR author: %s\n", err)
	}

	// Send the approval request message to the relay server.
	err = protocol.Write(c.ws, protocol.ApproveMessage{
		Link:   link,
		Author: author,
	})
	if err != nil {
		log.Printf("failed to write json message: %s", err)
	}
}
