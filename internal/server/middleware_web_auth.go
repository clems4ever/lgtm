package server

import (
	"context"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

// middlewareWebAuthMiddleware is an HTTP middleware that ensures GitHub OAuth authentication for web requests.
// It checks the session for a valid GitHub access token and username, and injects them into the request context.
// If authentication fails, it redirects the user to the OAuth2 login flow.
func (s *Server) middlewareWebAuthMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the session from the request cookies.
		session, err := s.sessionStore.Get(r, SessionName)
		if err != nil {
			log.Println("failed to get session", err)
			s.redirectAuth(w, r)
			return
		}

		// Extract the GitHub access token from the session values.
		accessToken, ok := session.Values[GhAccessTokenSessionKey].(string)
		if !ok {
			s.redirectAuth(w, r)
			return
		}

		// Extract the GitHub username from the session values.
		username, ok := session.Values[GhUsernameSessionKey].(string)
		if !ok {
			s.redirectAuth(w, r)
			return
		}

		// Add the username and access token to the request context for downstream handlers.
		ctx := context.WithValue(r.Context(), "username", username)
		ctx = context.WithValue(ctx, "access_token", accessToken)
		r = r.WithContext(ctx)

		// Call the wrapped handler with the updated request context.
		fn(w, r)
	}
}

// redirectAuth redirects the user to the GitHub OAuth2 login page to initiate authentication.
func (s *Server) redirectAuth(w http.ResponseWriter, r *http.Request) {
	authURL := s.oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, authURL, http.StatusFound)
}
