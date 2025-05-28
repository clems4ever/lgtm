package github

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// GetPRAuthor retrieves the GitHub username of the author of the given pull request.
// It fetches the PR metadata from the GitHub API using the provided access token.
//
// Parameters:
// - link: A PRLink representing the pull request.
//
// Returns:
// - The GitHub username of the PR author.
// - An error if the API request fails or the response cannot be parsed.
func (c *Client) GetPRAuthor(link PRLink) (string, error) {
	url := fmt.Sprintf("/repos/%s/%s/pulls/%d", link.Owner, link.Repo, link.PRNumber)
	resp, err := c.doNewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var pr struct {
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return "", err
	}
	return pr.User.Login, nil
}

// ApprovePR sends an approval review to the specified PR using the GitHub API.
// The approval includes the provided message as the review body.
//
// Parameters:
// - link: A PRLink representing the pull request.
// - message: The approval message to include in the review.
//
// Returns:
// - An error if the API request fails or the response indicates an error.
func (c *Client) ApprovePR(link PRLink, message string) error {
	url := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", link.Owner, link.Repo, link.PRNumber)
	body := fmt.Sprintf(`{"event":"APPROVE","body":%q}`, message)
	resp, err := c.doNewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s", string(data))
	}
	return nil
}

// IsPRAproved checks if the authenticated user has already approved the PR.
// It returns true if the latest review by the user is "APPROVED", false otherwise.
//
// Parameters:
// - link: A PRLink representing the pull request.
//
// Returns:
// - A boolean indicating whether the PR is approved by the user.
// - An error if the API request fails or the response cannot be parsed.
func (c *Client) IsPRAproved(link PRLink) (bool, error) {
	url := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", link.Owner, link.Repo, link.PRNumber)
	resp, err := c.doNewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("GitHub API error: %s", string(data))
	}
	type review struct {
		ID   int `json:"id"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
		State string `json:"state"`
	}
	var reviews []review
	if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
		return false, err
	}
	// Get the authenticated user's login
	userLogin, err := c.GetAuthenticatedUserLogin()
	if err != nil {
		return false, err
	}
	// Find the latest review by the user
	var latestReview *review
	for i := range reviews {
		r := &reviews[i]
		if r.User.Login == userLogin {
			if latestReview == nil || r.ID > latestReview.ID {
				latestReview = r
			}
		}
	}
	if latestReview != nil && latestReview.State == "APPROVED" {
		return true, nil
	}
	return false, nil
}
