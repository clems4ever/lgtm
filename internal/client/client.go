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

	// The local server that serves the callback and UI.
	localServer *LocalServer
	port        int

	// The path where the oauth2 token is stored on disk.
	githubTokenPath         string
	githubBaseURL           string
	githubAuthServerBaseURL string
	githubAuthCallback      github.AuthCallback

	// GitHub client for interacting with the GitHub API.
	githubClientMu sync.Mutex
	githubClient   *github.Client
	// HTTP client for making requests (can be overridden for testing).
	httpClient *http.Client
	// Mutex for synchronizing WebSocket access.
	wsMu sync.Mutex
	// WebSocket connection to the relay server.
	ws *websocket.Conn

	// Mutex for synchronizing access to available approvers.
	availableApproversMu sync.Mutex
	// List of available approvers.
	availableApprovers []string

	// Context for managing the lifecycle of the client.
	ctx  context.Context
	done context.CancelFunc
}

// NewClient creates a new instance of the lgtm client with the provided configuration.
func NewClient(
	serverURL string,
	authToken string,
	addr string,
	githubTokenPath string,
	githubBaseURL string,
	githubBaseURLAuthServer string,
	githubAuthCallback github.AuthCallback,
	httpClient *http.Client, // for testing
) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	port := extractPort(addr)

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		serverURL:               serverURL,
		authToken:               authToken,
		githubTokenPath:         githubTokenPath,
		githubBaseURL:           githubBaseURL,
		githubAuthServerBaseURL: githubBaseURLAuthServer,
		githubAuthCallback:      githubAuthCallback,
		httpClient:              httpClient,
		localServer:             NewLocalServer(addr),
		ctx:                     ctx,
		done:                    cancel,
		port:                    port,
	}
}

// Run starts the client: it launches the local HTTP server, authenticates with GitHub,
// and connects to the relay server via websocket, retrying on failure.
// It also gracefully shuts down the local server when the WebSocket connection is closed.
func (c *Client) Start() error {
	serverURL, err := url.Parse(c.serverURL)
	if err != nil {
		log.Fatal("failed to parse server url")
	}

	var wg sync.WaitGroup

	// Start the local HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.localServer.Start(c.githubAuthCallback, c)
	}()

	// Connect to the relay server via WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()

		c.autoconnectToWsServerAndListen(c.ctx, serverURL, c.authToken, func() {
			fmt.Printf("You can visit %s to submit your PRs\n", fmt.Sprintf("http://127.0.0.1:%d/", c.port))
		})

		// Shutdown the local server after WebSocket disconnect
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := c.localServer.Shutdown(ctx)
		if err != nil {
			log.Printf("failed to shutdown local server: %s\n", err)
		}
	}()

	wg.Wait()
	return nil
}

// Shutdown gracefully stops the client by canceling its context.
func (c *Client) Shutdown(ctx context.Context) error {
	c.done()
	return nil
}
