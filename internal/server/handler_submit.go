package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/clems4ever/lgtm/internal/github"
)

// SubmitBodyRequest represents the expected JSON body for a PR submission.
type SubmitBodyRequest struct {
	PRLink string `json:"pr_link"`
}

// handlerSubmit handles POST requests to submit a PR for approval.
// It parses the PR link, validates it, and attempts to forward it for approval.
// Returns appropriate HTTP status codes and messages for error/success cases.
func (s *Server) handlerSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var resp SubmitBodyRequest
	err := json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		log.Println("failed to decode body", err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	prLink, err := github.ParsePullRequestURL(resp.PRLink)
	if err != nil {
		log.Println("failed to parse pull request", err)
		http.Error(w, "Invalid pull request URL", http.StatusBadRequest)
		return
	}

	// Attempt to forward the PR for approval.
	err = s.RequestApproval(prLink)
	if err != nil {
		fmt.Printf("failed to approve PR %s: %s\n", prLink, err)
		if errors.Is(err, ErrNoEligibleApprover) {
			// No eligible approver found: return 422 Unprocessable Entity.
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// Internal error during approval process.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If successful, you may want to write a success message here.
}
