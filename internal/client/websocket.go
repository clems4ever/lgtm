package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/clems4ever/lgtm/internal/common"
	"github.com/clems4ever/lgtm/internal/github"
	"github.com/clems4ever/lgtm/internal/protocol"
	"github.com/gorilla/websocket"
)

var (
	// ErrUnauthorized is returned when the server responds with a 401 Unauthorized status.
	ErrUnauthorized = fmt.Errorf("unauthorized")

	// RetrySeconds defines the interval between retry attempts for WebSocket connections.
	RetrySeconds = 10 * time.Second
)

// autoconnectToWsServerAndListen continuously attempts to connect to the WebSocket server.
// If the connection fails, it retries after a specified interval unless the context is canceled.
//
// Parameters:
// - ctx: Context for managing the lifecycle of the connection attempts.
// - serverURL: The URL of the WebSocket server.
// - authToken: The authentication token to use for the connection.
// - connectionCallback: A callback function to execute upon successful connection.
func (c *Client) autoconnectToWsServerAndListen(ctx context.Context, serverURL *url.URL, authToken string, connectionCallback func()) {
	// Retry connection every X seconds if it fails
	for {
		err := c.connectToWsServerAndListen(ctx, serverURL, authToken, connectionCallback)
		if err != nil {
			if err == ErrUnauthorized {
				log.Println(err)
				return
			}
			if ctx.Err() != nil {
				return
			}
			log.Printf("%v. Retrying in %s...", err, RetrySeconds)
			select {
			case <-time.After(RetrySeconds):
			case <-ctx.Done():
				return
			}
			continue
		}
		break // exit loop if registerWithWSServer returns nil (clean disconnect)
	}
}

// getServerConfig retrieves the server configuration, including OAuth app details, from the server.
//
// Parameters:
// - ctx: Context for managing the lifecycle of the HTTP request.
// - serverURL: The base URL of the server.
// - authToken: The authentication token to use for the request.
//
// Returns:
// - A common.SharedConfig containing the server configuration.
// - An error if the request fails or the response cannot be parsed.
func (c *Client) getServerConfig(ctx context.Context, serverURL *url.URL, authToken string) (common.SharedConfig, error) {
	// Retrieve the servernfiguration for OAuth details
	configURL := *serverURL
	configURL.Path = "/config"
	req, err := http.NewRequestWithContext(ctx, "GET", configURL.String(), nil)
	if err != nil {
		return common.SharedConfig{}, fmt.Errorf("failed to create config request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return common.SharedConfig{}, fmt.Errorf("failed to retrieve server config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.SharedConfig{}, fmt.Errorf("failed to retrieve server config, status: %d", resp.StatusCode)
	}

	var serverConfig common.SharedConfig
	if err := json.NewDecoder(resp.Body).Decode(&serverConfig); err != nil {
		return common.SharedConfig{}, fmt.Errorf("failed to decode server config: %w", err)
	}
	return serverConfig, nil

}

// connectToWsServerAndListen establishes a WebSocket connection to the relay server.
// It retrieves the server configuration, authenticates with GitHub, registers the client as an approver,
// and listens for approval requests.
//
// Parameters:
// - ctx: Context for managing the lifecycle of the connection.
// - serverURL: The URL of the WebSocket server.
// - authToken: The authentication token to use for the connection.
// - connectionCallback: A callback function to execute upon successful connection.
//
// Returns:
// - An error if the connection fails or the server responds with an error.
func (c *Client) connectToWsServerAndListen(ctx context.Context, serverURL *url.URL, authToken string, connectionCallback func()) error {
	serverConfig, err := c.getServerConfig(ctx, serverURL, authToken)
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	// Build OAuth2 configuration using server details
	oauthConfig := OauthConfigBuilder(OAuthConfigBuilderArgs{
		AuthServerBaseURL: c.githubAuthServerBaseURL,
		ClientID:          serverConfig.ClientID,
		ClientSecret:      serverConfig.ClientSecret,
		Port:              c.port,
	})

	// Authenticate with GitHub
	githubClient, err := github.Authenticate(
		c.ctx,
		c.githubTokenPath,
		oauthConfig,
		c.githubBaseURL,
		c.githubAuthCallback,
		c.httpClient)
	if err != nil {
		return fmt.Errorf("failed to authenticate against Github: %w", err)
	}
	c.githubClient = githubClient

	// Connect to WebSocket server
	var wsURL = *serverURL
	wsURL.Scheme = "ws"
	wsURL.Path = "/ws"

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fmt.Printf("connecting to server at %s... ", wsURL.String())
	headers := make(map[string][]string)
	if authToken != "" {
		headers["Authorization"] = []string{"Bearer " + authToken}
	}
	conn, res, err := websocket.DefaultDialer.DialContext(ctx, wsURL.String(), headers)
	if err != nil {
		if res != nil {
			if res.StatusCode == 401 {
				return ErrUnauthorized
			}
			return fmt.Errorf("server replied with %d", res.StatusCode)
		}
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	fmt.Println("connected!")

	c.wsMu.Lock()
	c.ws = conn
	c.wsMu.Unlock()

	// Register the approver against the server.
	err = c.registerApprover(conn, connectionCallback)
	if err != nil {
		return fmt.Errorf("failed to register approver: %w", err)
	}

	// Listen for PR approval requests
	for {
		var msg protocol.Message
		if err := protocol.Read(conn, &msg); err != nil {
			break
		}

		switch v := msg.Message.(type) {
		case protocol.ApproveMessage:
			err = c.handleApproveMessage(v)
			if err != nil {
				log.Printf("failed to handle message: %s\n", err)
			}
		case protocol.ApproverEventMessage:
			c.availableApproversMu.Lock()
			c.availableApprovers = v.Approvers
			c.availableApproversMu.Unlock()
		default:
			log.Println("unsupported message type")
			continue
		}
	}

	log.Println("disconnected from server")
	return fmt.Errorf("disconnected")
}

// handleApproveMessage processes an ApproveMessage received from the relay server.
// It attempts to approve the pull request if it has not already been approved by the user.
//
// Parameters:
// - msg: The ApproveMessage containing the pull request details.
//
// Returns:
// - An error if the approval process fails.
func (c *Client) handleApproveMessage(msg protocol.ApproveMessage) error {
	// alreadyApproved, err := c.githubClient.IsPRAproved(msg.Link)
	// if err != nil {
	// 	return fmt.Errorf("failed to check PR approval: %w", err)
	// }
	// if alreadyApproved {
	// 	log.Printf("PR %s already approved by this user", msg.Link)
	// 	return nil
	// }
	err := c.githubClient.ApprovePR(msg.Link, "lgtm")
	if err != nil {
		return fmt.Errorf("failed to approve PR: %w", err)
	}
	log.Printf("✅ PR %s approved successfully!", msg.Link)
	return nil
}

// registerApprover registers the client as an approver for its repositories with the server.
//
// Parameters:
// - conn: The WebSocket connection to the server.
// - connectionCallback: A callback function to execute upon successful registration.
//
// Returns:
// - An error if the registration process fails.
func (c *Client) registerApprover(conn *websocket.Conn, connectionCallback func()) error {
	// Retrieve the list of repos this client can approve using the GitHub token.
	repos, err := c.githubClient.GetRepos()
	if err != nil {
		return fmt.Errorf("failed to retrieve repos from github: %w", err)
	}

	userLogin, err := c.githubClient.GetAuthenticatedUserLogin()
	if err != nil {
		return fmt.Errorf("failed to retrieve user login: %w", err)
	}

	fmt.Println("You are registered as approver for the following repositories:")
	for _, repo := range repos {
		fmt.Println("- " + repo)
	}
	fmt.Println()

	connectionCallback()

	reg := protocol.RegisterRequestMessage{
		Repos:      repos,
		GithubUser: userLogin,
	}
	if err := protocol.Write(conn, reg); err != nil {
		return fmt.Errorf("failed to write json message: %w", err)
	}

	return nil
}
