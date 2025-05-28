package github

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// PRLink represents a GitHub pull request link with owner, repo, and PR number.
// It provides methods to retrieve the full repository name and the canonical PR URL.
type PRLink struct {
	Owner    string // Repository owner
	Repo     string // Repository name
	PRNumber int    // Pull request number
}

// RepoFullName returns the "owner/repo" string for the PR.
// This is useful for identifying the repository in a concise format.
func (l *PRLink) RepoFullName() string {
	return fmt.Sprintf("%s/%s", l.Owner, l.Repo)
}

// String returns the canonical GitHub PR URL for this PRLink.
// The URL is constructed using the owner, repo, and PR number.
func (l PRLink) String() string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", l.Owner, l.Repo, l.PRNumber)
}

// ParsePullRequestURL extracts owner, repo, and PR number from a GitHub PR URL.
// It validates the URL format and ensures it corresponds to a pull request.
//
// Parameters:
// - link: The GitHub PR URL to parse.
//
// Returns:
// - A PRLink containing the owner, repo, and PR number.
// - An error if the URL is invalid or does not match the expected PR format.
func ParsePullRequestURL(link string) (PRLink, error) {
	u, err := url.Parse(link)
	if err != nil {
		return PRLink{}, fmt.Errorf("invalid URL: %w", err)
	}
	parts := strings.Split(u.Path, "/")
	// Expect: /{owner}/{repo}/pull/{number}
	if len(parts) < 5 || parts[3] != "pull" {
		return PRLink{}, fmt.Errorf("invalid PR link format")
	}

	owner := parts[1]
	repo := parts[2]
	prNumber, err := strconv.Atoi(parts[4])
	if err != nil {
		return PRLink{}, fmt.Errorf("invalid PR number: %w", err)
	}

	prLink := PRLink{
		Owner:    owner,
		Repo:     repo,
		PRNumber: prNumber,
	}
	return prLink, nil
}
