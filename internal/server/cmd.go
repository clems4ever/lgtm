package server

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/clems4ever/lgtm/internal/common"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	addrFlag         string
	authTokenFlag    string
	clientIDFlag     string
	clientSecretFlag string
)

const (
	defaultAddr = ":8080"
)

type clientInfo struct {
	conn *websocket.Conn
	// if the githubUser variable is not set, it means the connection is established but
	// the client have not registered yet.
	githubUser string
	repos      map[string]struct{} // set of "owner/repo"
}

var (
	upgrader      = websocket.Upgrader{}
	clients       = make(map[*websocket.Conn]*clientInfo)
	clientsByRepo = make(map[string][]*clientInfo)
	users         = make(map[string]struct{})
	mu            sync.Mutex
)

// authMiddleware wraps an HTTP handler with authentication logic.
// It checks the Authorization header for a Bearer token and ensures it matches the provided authToken.
// If the token is invalid, the request is rejected with a 401 Unauthorized status.
func authMiddleware(authToken string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate using the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+authToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		fn(w, r)
	}
}

// BuildCommand creates the Cobra command for running the server.
// It sets up the WebSocket handler with optional authentication based on the provided auth token.
func BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Runs the server",
		Run: func(cmd *cobra.Command, args []string) {
			var server = Server{
				cfg: common.SharedConfig{
					ClientID:     clientIDFlag,
					ClientSecret: clientSecretFlag,
				},
			}
			var wsHandlerFn = server.wsHandler
			var configHandlerFn = server.configHandler
			// Wrap the WebSocket handler with authentication middleware if an auth token is provided
			if authTokenFlag != "" {
				wsHandlerFn = authMiddleware(authTokenFlag, wsHandlerFn)
				configHandlerFn = authMiddleware(authTokenFlag, configHandlerFn)
			} else {
				log.Println("WARN no auth token provided, the server is not secured")
			}
			http.HandleFunc("/ws", wsHandlerFn)
			http.HandleFunc("/config", configHandlerFn)
			log.Printf("Server listening on %s\n", addrFlag)
			log.Fatal(http.ListenAndServe(addrFlag, nil))
		},
	}

	cmd.Flags().StringVar(&addrFlag, "addr", defaultAddr, "addr to listen on")
	cmd.Flags().StringVar(&authTokenFlag, "auth-token", "", "shared authentication token")
	cmd.Flags().StringVar(&clientIDFlag, "client-id", os.Getenv("GITHUB_CLIENT_ID"), "client id of the gh app")
	cmd.Flags().StringVar(&clientSecretFlag, "client-secret", os.Getenv("GITHUB_CLIENT_SECRET"), "client secret of the gh app")
	return cmd
}
