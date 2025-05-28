package client

import (
	"log"
	"net"
	"strconv"

	"github.com/clems4ever/lgtm/internal/github"
	"github.com/spf13/cobra"
)

var (
	serverURLFlag     string
	addrFlag          string
	authTokenFlag     string
	authServerURLFlag string
)

const (
	defaultServerURL     = "https://lgtm.clems4ever.com"
	defaultAuthServerURL = "https://github.com/login/oauth"
	defaultAddr          = ":8081"
)

// extractPort extracts the port number from a given address string.
// It expects the address to be in the format "host:port" and returns the port as an integer.
// If the address is invalid or the port cannot be parsed, the function logs a fatal error.
func extractPort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatalf("failed to parse port: %s", err)
	}
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		log.Fatalf("failed to parse int: %s", err)
	}
	return int(port)
}

// BuildCommand creates the root Cobra command for the lgtm client.
// It sets up flags for configuring the server URL, local address, and authentication token.
// The command starts the client, which includes launching the local server, authenticating with GitHub,
// and connecting to the relay server via WebSocket.
func BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client commands for lgtm",
		Run: func(cmd *cobra.Command, args []string) {
			// Retrieve the path to the GitHub token file
			path, err := github.GetTokenFilePath()
			if err != nil {
				log.Fatalf("failed to get token file path: %s", err)
			}

			// Initialize the browser-based authentication callback
			authCallback := NewBrowserAuthCallback()

			// Start the client with the provided configuration
			err = NewClient(serverURLFlag, authTokenFlag, addrFlag, path, "", authServerURLFlag, authCallback, nil).
				Start()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	// Define flags for the command
	cmd.Flags().StringVar(&serverURLFlag, "server-url", defaultServerURL, "url to the lgtm relay server")
	cmd.Flags().StringVar(&addrFlag, "addr", defaultAddr, "addr to listen on")
	cmd.Flags().StringVar(&authTokenFlag, "auth-token", "", "shared authentication token")
	cmd.Flags().StringVar(&authServerURLFlag, "auth-server-url", defaultAuthServerURL, "url to the GitHub OAuth server")

	return cmd
}
