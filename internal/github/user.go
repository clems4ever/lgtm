package github

import (
	"encoding/json"
	"fmt"
	"io"
)

// GetAuthenticatedUserLogin fetches the login (username) of the authenticated user from GitHub.
// It uses the provided access token to authenticate the request.
func (c *Client) GetAuthenticatedUserLogin() (string, error) {
	resp, err := c.doNewRequest("GET", "/user", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error: %s", string(data))
	}
	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}
	return user.Login, nil
}
