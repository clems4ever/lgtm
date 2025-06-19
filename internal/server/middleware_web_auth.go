package server

import (
	"context"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

// middlewareWebAuthMiddleware wraps an HTTP handler with authentication logic.
// It verifies the user's session by checking the session cookie and retrieving the GitHub access token.
// If the session is valid, it fetches the authenticated user's GitHub username and adds it to the request context.
// If the session is invalid or the user cannot be authenticated, the user is redirected to the OAuth2 login flow.
func (s *Server) middlewareWebAuthMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the session
		session, err := s.sessionStore.Get(r, SessionName)
		if err != nil {
			log.Println("failed to get session", err)
			s.redirectAuth(w, r)
			return
		}

		// Extract the access token and username from the session
		accessToken, ok := session.Values[GhAccessTokenSessionKey].(string)
		if !ok {
			s.redirectAuth(w, r)
			return
		}

		username, ok := session.Values[GhUsernameSessionKey].(string)
		if !ok {
			s.redirectAuth(w, r)
			return
		}

		// Add the username and access token to the request context
		ctx := context.WithValue(r.Context(), "username", username)
		ctx = context.WithValue(ctx, "access_token", accessToken)
		r = r.WithContext(ctx)

		// Call the wrapped handler
		fn(w, r)
	}
}

func (s *Server) redirectAuth(w http.ResponseWriter, r *http.Request) {
	authURL := s.oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, authURL, http.StatusFound)
}
