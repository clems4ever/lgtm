package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/clems4ever/lgtm/internal/github"
	"github.com/gorilla/websocket"
)

// Client represents a lgtm client instance, including its configuration, GitHub authentication,
// local server, and WebSocket connection to the relay server.
type Client struct {
	// The URL to the lgtm relay server.
	serverURL string
	// The shared authentication token for the relay server.
	authToken string

	// GitHub client for interacting with the GitHub API.
	githubClient *github.Client
	// The handle of the github user this client is served with.
	githubUsername    string
	reconnectInterval time.Duration
	// Mutex for synchronizing WebSocket access.
	wsMu sync.Mutex

	// WebSocket connection to the relay server.
	ws *websocket.Conn

	// Context for managing the lifecycle of the client.
	ctx  context.Context
	done context.CancelFunc
}

// NewClient creates a new instance of the lgtm client with the provided configuration.
// githubToken must be a valid GitHub token with 'repo' and 'read:user' permissions.
func NewClient(
	serverURL string,
	authToken string,
	reconnectInterval time.Duration,
	githubToken string,
	githubAPIBaseURL string,
	httpClient *http.Client, // for testing
) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	// Create a GitHub client using the provided token and HTTP client.
	ghClient := github.NewClient(githubToken, githubAPIBaseURL, httpClient)

	// Retrieve the authenticated GitHub username.
	ghUsername, err := ghClient.GetAuthenticatedUserLogin()
	if err != nil {
		return nil, fmt.Errorf("failed to get username: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		serverURL:         serverURL,
		authToken:         authToken,
		reconnectInterval: reconnectInterval,
		ctx:               ctx,
		done:              cancel,
		githubClient:      ghClient,
		githubUsername:    ghUsername,
	}, nil
}

// Start launches the client: it connects to the relay server via websocket and listens for events.
// It blocks until the client is stopped.
func (c *Client) Start() error {
	serverURL, err := url.Parse(c.serverURL)
	if err != nil {
		log.Fatal("failed to parse server url")
	}

	var wg sync.WaitGroup

	// Connect to the relay server via WebSocket in a goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.autoconnectToWsServerAndListen(c.ctx, serverURL, c.authToken)
	}()

	wg.Wait()
	return nil
}

// Stop gracefully stops the client by canceling its context.
func (c *Client) Stop() {
	c.done()
}
