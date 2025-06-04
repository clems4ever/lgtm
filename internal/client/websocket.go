package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/clems4ever/lgtm/internal/protocol"
	"github.com/gorilla/websocket"
)

var (
	// ErrUnauthorized is returned when the server responds with a 401 Unauthorized status.
	ErrUnauthorized = fmt.Errorf("unauthorized")
)

// autoconnectToWsServerAndListen continuously attempts to connect to the WebSocket server.
// If the connection fails, it retries after a specified interval unless the context is canceled.
func (c *Client) autoconnectToWsServerAndListen(ctx context.Context, serverURL *url.URL, authToken string) {
	// Retry connection every X seconds if it fails
	for {
		err := c.connectToWsServerAndListen(ctx, serverURL, authToken)
		if err != nil {
			if err == ErrUnauthorized {
				log.Println(err)
				return
			}
			if ctx.Err() != nil {
				return
			}
			fmt.Printf("%v. Retrying in %s...\n", err, c.reconnectInterval)
			select {
			case <-time.After(c.reconnectInterval):
			case <-ctx.Done():
				return
			}
			continue
		}
		break // exit loop if registerWithWSServer returns nil (clean disconnect)
	}
}

// connectToWsServerAndListen establishes a WebSocket connection to the relay server.
// It retrieves the server configuration, authenticates with GitHub, registers the client as an approver,
// and listens for approval requests.
func (c *Client) connectToWsServerAndListen(ctx context.Context, serverURL *url.URL, authToken string) error {
	// Prepare the WebSocket URL (ws or wss)
	var wsURL = *serverURL
	if serverURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}
	wsURL.Path = "/ws"

	// Set a timeout for the connection attempt
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
			json.NewDecoder(os.Stdout).Decode(res.Body)
			return fmt.Errorf("server replied with %d", res.StatusCode)
		}
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer res.Body.Close()
	fmt.Println("connected!")

	// Store the WebSocket connection
	c.wsMu.Lock()
	c.ws = conn
	c.wsMu.Unlock()

	// Register the approver against the server.
	err = c.registerApprover(conn)
	if err != nil {
		return fmt.Errorf("failed to register approver: %w", err)
	}

	var wg sync.WaitGroup
	// this channel allows to signal the ping routine that the connection has closed.
	connectionClosedC := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Listen for PR approval requests from the server
		for {
			var msg protocol.Message
			if err := protocol.Read(conn, &msg); err != nil {
				break
			}

			switch v := msg.Message.(type) {
			case protocol.ApproveRequestMessage:
				// Handle an approval request message
				err = c.handleApproveMessage(conn, msg.RequestID, v)
				if err != nil {
					log.Printf("failed to handle message: %s\n", err)
				}
			default:
				log.Println("unsupported message type", msg.Type)
				continue
			}
		}
		log.Println("disconnected from server")
		close(connectionClosedC)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				_, err := protocol.Write(conn, protocol.PingMessage{})
				if err != nil {
					log.Println("failed to ping")
				}
			case <-connectionClosedC:
				return
			}
		}
	}()

	wg.Wait()
	return fmt.Errorf("disconnected")
}

// handleApproveMessage processes an ApproveRequestMessage received from the relay server.
// It attempts to approve the pull request if it has not already been approved by the user.
func (c *Client) handleApproveMessage(conn *websocket.Conn, reqID string, msg protocol.ApproveRequestMessage) error {
	author, err := c.githubClient.GetPRAuthor(msg.Link)
	if err != nil {
		return fmt.Errorf("failed to get PR author: %w", err)
	}

	// If the author is the same as the current user, respond with an error
	if author == c.githubUsername {
		err = protocol.WriteWithRequestID(conn, protocol.ApproveResponseMessage{
			Response: protocol.ApproveResponseErrSameAuthor,
		}, reqID)
		if err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
		return nil
	}

	// Optionally, check if already approved (commented out)
	// alreadyApproved, err := c.githubClient.IsPRAproved(msg.Link)
	// if err != nil {
	// 	return fmt.Errorf("failed to check PR approval: %w", err)
	// }
	// if alreadyApproved {
	// 	log.Printf("PR %s already approved by this user", msg.Link)
	// 	return nil
	// }

	// Attempt to approve the PR
	err = c.githubClient.ApprovePR(msg.Link, "lgtm")
	if err != nil {
		return fmt.Errorf("failed to approve PR: %w", err)
	}

	// Respond with success
	err = protocol.WriteWithRequestID(conn, protocol.ApproveResponseMessage{
		Response: protocol.ApproveResponseSuccess,
	}, reqID)
	if err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}
	log.Printf("✅ PR %s approved successfully!", msg.Link)
	return nil
}

// registerApprover registers the client as an approver for its repositories with the server.
// It retrieves the list of repos this client can approve using the GitHub token and sends a registration message.
func (c *Client) registerApprover(conn *websocket.Conn) error {
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

	reg := protocol.RegisterRequestMessage{
		Repos:      repos,
		GithubUser: userLogin,
	}
	if _, err := protocol.Write(conn, reg); err != nil {
		return fmt.Errorf("failed to write json message: %w", err)
	}

	return nil
}
