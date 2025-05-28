package github

import (
	"encoding/json"
	"fmt"
	"io"
)

type ReposResponseBody []Repo

type RepoPermissions struct {
	Push bool `json:"push"`
}

type Repo struct {
	FullName    string          `json:"full_name"`
	Permissions RepoPermissions `json:"permissions"`
}

// GetRepos returns the list of repositories the authenticated user can approve PRs for.
// Only repositories where the user has push access (i.e., can approve PRs) are included.
// It uses the GitHub API and the user's OAuth token.
func (c *Client) GetRepos() ([]string, error) {
	resp, err := c.doNewRequest("GET", "/user/repos?per_page=100", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(data))
	}

	var repos ReposResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	var result []string
	for _, repo := range repos {
		// Only include repos where the user has push access (can approve PRs)
		if repo.Permissions.Push {
			result = append(result, repo.FullName)
		}
	}
	return result, nil
}
