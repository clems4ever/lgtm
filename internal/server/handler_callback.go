package server

import (
	"log"
	"net/http"

	"github.com/clems4ever/lgtm/internal/github"
)

func (s *Server) handlerCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	// Extract the authorization code from the query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for an access token
	token, err := s.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("failed to exchange authorization code: %v, code: %s", err, code)
		http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
		return
	}

	session, err := s.sessionStore.Get(r, SessionName)
	if err != nil {
		log.Printf("failed to get session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	session.Values[GhAccessTokenSessionKey] = token.AccessToken
	gh := github.NewClient(token.AccessToken, defaultGithubAPIURL, s.httpClient)

	username, err := gh.GetAuthenticatedUserLogin()
	if err != nil {
		log.Printf("failed to get user login: %v, token: %s", err, token.AccessToken)
		http.Error(w, "Failed to retrieve user information", http.StatusInternalServerError)
		return
	}

	session.Values[GhUsernameSessionKey] = username

	err = session.Save(r, w)
	if err != nil {
		log.Printf("failed to save session: %v, username: %s", err, username)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect the user to the home page
	http.Redirect(w, r, "/", http.StatusFound)
}
