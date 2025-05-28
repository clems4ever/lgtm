package github

import (
	"io"
	"net/http"
	"time"
)

const defaultAPIBaseURL = "https://api.github.com"

// Client represents a GitHub API client, handling authentication and requests.
// It uses an access token for authorization and supports customizable HTTP clients and API base URLs.
type Client struct {
	httpClient  *http.Client // HTTP client used for requests
	accessToken string       // GitHub OAuth access token
	apiBaseURL  string       // Base URL for GitHub API (e.g., "https://api.github.com")
}

// NewClient creates a new GitHub API client with the given access token, API base URL, and optional HTTP client.
// If apiBaseURL is empty, "https://api.github.com" is used.
// If httpClient is nil, http.DefaultClient is used.
//
// Parameters:
// - accessToken: GitHub OAuth access token for authorization.
// - apiBaseURL: Base URL for GitHub API requests.
// - httpClient: Optional HTTP client for making requests.
//
// Returns:
// - A new instance of the GitHub API client.
func NewClient(accessToken string, apiBaseURL string, httpClient *http.Client) *Client {
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient:  httpClient,
		accessToken: accessToken,
		apiBaseURL:  apiBaseURL,
	}
}

// newRequest creates a new HTTP request with the correct base URL and authorization headers.
//
// Parameters:
// - method: HTTP method (e.g., "GET", "POST").
// - path: API endpoint path (e.g., "/repos").
// - body: Optional request body.
//
// Returns:
// - An HTTP request object.
// - An error if the request creation fails.
func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.apiBaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+c.accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	return req, nil
}

// doRequest executes an HTTP request using the client's httpClient or a default one.
//
// Parameters:
// - req: The HTTP request to execute.
//
// Returns:
// - An HTTP response object.
// - An error if the request execution fails.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	client := c.httpClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return client.Do(req)
}

// doNewRequest creates and executes an HTTP request in one step.
//
// Parameters:
// - method: HTTP method (e.g., "GET", "POST").
// - path: API endpoint path (e.g., "/repos").
// - body: Optional request body.
//
// Returns:
// - An HTTP response object.
// - An error if the request creation or execution fails.
func (c *Client) doNewRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}
