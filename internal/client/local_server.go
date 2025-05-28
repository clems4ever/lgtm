package client

import (
	"context"
	"log"
	"net/http"

	"github.com/clems4ever/lgtm/internal/github"
)

// LocalServer wraps the HTTP server logic for the client.
// It manages the HTTP server, its routes, and the lifecycle of the server.
type LocalServer struct {
	codeCh chan<- string // Channel for passing authentication codes

	mux    *http.ServeMux // HTTP request multiplexer
	server *http.Server   // HTTP server instance
}

// NewLocalServer creates a new Server instance with the given address.
// It initializes the HTTP server and multiplexer.
//
// Parameters:
// - addr: The address on which the server will listen.
//
// Returns:
// - A new Server instance.
func NewLocalServer(addr string) *LocalServer {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return &LocalServer{
		mux:    mux,
		server: server,
	}
}

// Start launches the HTTP server and registers handlers for various routes.
// It handles the home page, PR approval, and OAuth callback.
//
// Parameters:
// - receiver: An AuthCodeReceiver for handling authentication codes.
// - client: The Client instance to associate with the server.
func (s *LocalServer) Start(receiver github.AuthCodeReceiver, client *Client) {
	s.mux.HandleFunc("/", client.handlerHome)
	s.mux.HandleFunc("/approve", client.handlerApprove)
	s.mux.HandleFunc("/callback", HandlerCallback(receiver))

	err := s.server.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Shutdown gracefully stops the HTTP server.
//
// Parameters:
// - ctx: Context for managing the shutdown process.
//
// Returns:
// - An error if the shutdown fails.
func (s *LocalServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
