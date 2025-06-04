package client

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	serverURLFlag         string
	reconnectIntervalFlag time.Duration
	pingIntervalFlag      time.Duration
)

const (
	defaultServerURL         = "https://lgtm.clems4ever.com"
	defaultReconnectInterval = 15 * time.Second
	defaultPingInterval      = 10 * time.Second
)

// BuildCommand creates the root Cobra command for the lgtm client.
// It sets up flags for configuring the server URL, local address, and authentication token.
// The command starts the client, which includes launching the local server, authenticating with GitHub,
// and connecting to the relay server via WebSocket.
func BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client commands for lgtm",
		Run: func(cmd *cobra.Command, args []string) {
			githubToken := os.Getenv("LGTM_GITHUB_TOKEN")
			if githubToken == "" {
				fmt.Println("LGTM_GITHUB_TOKEN env var must be provided. " +
					"Make sure it has the 'repo' and 'read:user' permissions and that the token is authorized " +
					"on all orgs you want to be an approver for.")
				os.Exit(1)
			}

			authToken := os.Getenv("LGTM_API_AUTH_TOKEN")

			// Start the client with the provided configuration
			c, err := NewClient(
				serverURLFlag,
				authToken,
				reconnectIntervalFlag,
				pingIntervalFlag,
				githubToken, "", nil)
			if err != nil {
				log.Fatal(err)
			}

			err = c.Start()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	// Define flags for the command
	cmd.Flags().StringVar(&serverURLFlag, "server-url", defaultServerURL, "url to the lgtm relay server")
	cmd.Flags().DurationVar(&reconnectIntervalFlag, "reconnect-interval", defaultReconnectInterval, "time between two reconnection attempts")
	cmd.Flags().DurationVar(&pingIntervalFlag, "ping-interval", defaultPingInterval, "interval for websocket ping messages")

	return cmd
}
