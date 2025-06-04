package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/clems4ever/lgtm/internal/common"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/spf13/cobra"
)

// Embed the static assets directory for serving files like favicon, manifest, etc.
//
//go:embed ui/assets/*
var staticAssets embed.FS

var (
	addrFlag          string // HTTP listen address
	baseURLFlag       string // Base URL for OAuth2 redirect
	authServerURLFlag string
	pingIntervalFlag  time.Duration
)

const (
	defaultAddr          = ":8080"
	defaultAuthServerURL = "https://github.com/login/oauth"
	defaultBaseURL       = "https://lgtm.clems4evever.com"
	defaultGithubAPIURL  = "https://api.github.com"
	defaultPingInterval  = 10 * time.Second
)

// BuildCommand creates the Cobra command for running the server.
// It sets up the HTTP server, session store, static file serving, and all routes.
func BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Runs the server",
		Run: func(cmd *cobra.Command, args []string) {
			// Read secrets from environment variables
			apiAuthToken := os.Getenv("LGTM_API_AUTH_TOKEN")
			clientID := os.Getenv("LGTM_GITHUB_CLIENT_ID")
			if clientID == "" {
				log.Fatal("LGTM_GITHUB_CLIENT_ID must be set")
			}
			clientSecret := os.Getenv("LGTM_GITHUB_CLIENT_SECRET")
			if clientSecret == "" {
				log.Fatal("LGTM_GITHUB_CLIENT_SECRET must be set")
			}
			sessionStoreEncryptionKey := os.Getenv("LGTM_SESSION_STORE_ENCRYPTION_KEY")
			if sessionStoreEncryptionKey == "" {
				log.Fatal("LGTM_SESSION_STORE_ENCRYPTION_KEY must be set")
			}

			// Initialize the main server struct with OAuth2 config
			var server = NewServer(
				common.OauthConfigBuilder(common.OAuthConfigBuilderArgs{
					AuthServerBaseURL: authServerURLFlag,
					ClientID:          clientID,
					ClientSecret:      clientSecret,
					Scopes:            []string{"read:user"},
					RedirectURL:       baseURLFlag + "/callback",
				}), pingIntervalFlag)
			defer server.Close()

			// Initialize the session store for secure cookie-based sessions
			cookieStore := sessions.NewCookieStore([]byte(sessionStoreEncryptionKey))
			cookieStore.MaxAge(3600) // Set session expiration to 1 hour
			cookieStore.Options.Path = "/"
			cookieStore.Options.HttpOnly = true
			cookieStore.Options.Secure = true // Ensure cookies are sent over HTTPS
			server.sessionStore = cookieStore

			// Create a new router for all HTTP routes
			router := mux.NewRouter()

			// Serve static assets (e.g., favicon, manifest, icons) under /assets/
			subFS, err := fs.Sub(staticAssets, "ui/assets")
			if err != nil {
				log.Fatalf("failed to create sub filesystem: %v", err)
			}
			staticHandler := http.FileServer(http.FS(subFS))
			router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", staticHandler))

			// Define application routes with appropriate middleware
			router.HandleFunc("/", server.middlewareWebAuthMiddleware(server.handlerHome)).Methods(http.MethodGet)
			router.HandleFunc("/submit", server.middlewareWebAuthMiddleware(server.handlerSubmit)).Methods(http.MethodPost)
			router.HandleFunc("/callback", server.handlerCallback).Methods(http.MethodGet)
			router.HandleFunc("/ws", apiAuthMiddleware(apiAuthToken, server.wsHandler)).Methods(http.MethodGet)

			// Custom 404 handler for undefined paths
			router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Resource not found", http.StatusNotFound)
			})

			log.Printf("Server listening on %s\n", addrFlag)
			log.Fatal(http.ListenAndServe(addrFlag, router))
		},
	}

	// Define command-line flags for server configuration
	cmd.Flags().StringVar(&addrFlag, "addr", defaultAddr, "addr to listen on")
	cmd.Flags().StringVar(&baseURLFlag, "base-url", defaultBaseURL, "base URL of the service being served (for oauth2 redirect)")
	cmd.Flags().StringVar(&authServerURLFlag, "auth-server-url", defaultAuthServerURL, "url to the GitHub OAuth server")
	cmd.Flags().DurationVar(&pingIntervalFlag, "ping-interval", defaultPingInterval, "interval for websocket ping messages")
	return cmd
}
