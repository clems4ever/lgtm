package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"golang.org/x/oauth2"
)

// RepoPermissions represents permissions for a repo.
type RepoPermissions struct {
	Push bool `json:"push"`
}

// Repo represents a GitHub repository for mocking.
type Repo struct {
	FullName    string          `json:"full_name"`
	Permissions RepoPermissions `json:"permissions"`
}

// GithubMockServer is a test server that mocks both GitHub API and OAuth2 endpoints.
type GithubMockServer struct {
	t *testing.T

	mu         sync.Mutex
	users      map[string]string // accessToken -> username
	repos      map[string][]Repo // username -> []Repo
	oauthCodes map[string]string // code -> username

	server     *httptest.Server
	oauth2Conf *oauth2.Config
}

// NewGithubMockServer creates a new GithubMockServer.
func NewGithubMockServer(t *testing.T, callbackURL string) *GithubMockServer {
	g := &GithubMockServer{
		t:          t,
		users:      make(map[string]string),
		repos:      make(map[string][]Repo),
		oauthCodes: make(map[string]string),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/user", g.handleUser)
	mux.HandleFunc("/user/repos", g.handleUserRepos)
	mux.HandleFunc("/repos/", g.handleRepoPR)
	mux.HandleFunc("/authorize", g.handleAuth)
	mux.HandleFunc("/access_token", g.handleToken)
	g.server = httptest.NewServer(mux)

	g.oauth2Conf = &oauth2.Config{
		ClientID:     "testclientid",
		ClientSecret: "testsecret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  g.server.URL + "/authorize",
			TokenURL: g.server.URL + "/access_token",
		},
		RedirectURL: callbackURL,
		Scopes:      []string{"repo"},
	}
	return g
}

// AddUser adds a user with a username and associates it with a repos list.
func (g *GithubMockServer) AddUser(username, accessToken string, repos []Repo) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.users[accessToken] = username
	g.repos[username] = repos
}

// AddOAuthCode associates an auth code with a username.
func (g *GithubMockServer) AddOAuthCode(code, username string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.oauthCodes[code] = username
}

// OAuth2Config returns the oauth2.Config for this mock server.
func (g *GithubMockServer) OAuth2Config() *oauth2.Config {
	return g.oauth2Conf
}

// URL returns the base URL of the mock server.
func (g *GithubMockServer) URL() string {
	return g.server.URL
}

// Close shuts down the mock server.
func (g *GithubMockServer) Close() {
	g.server.Close()
}

// --- Handlers ---

func (g *GithubMockServer) handleUser(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	g.mu.Lock()
	username := g.users[token]
	g.mu.Unlock()
	if username == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"login":"%s"}`, username)
}

func (g *GithubMockServer) handleUserRepos(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	g.mu.Lock()
	username := g.users[token]
	repos := g.repos[username]
	g.mu.Unlock()
	if username == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(repos)
}

func (g *GithubMockServer) handleRepoPR(w http.ResponseWriter, r *http.Request) {
	// Example: /repos/{owner}/{repo}/pulls/{number}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 6 && parts[3] != "" && parts[4] != "" && parts[5] == "pulls" {
		// Simulate PR author as "prauthor"
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"user":{"login":"prauthor"}}`)
		return
	}
	http.NotFound(w, r)
}

func (g *GithubMockServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	// Simulate user login and redirect with code
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "mockuser"
	}
	redirectURI := r.URL.Query().Get("redirect_uri")
	code := "code-" + username
	g.AddOAuthCode(code, username)
	http.Redirect(w, r, redirectURI+"?code="+code, http.StatusFound)
}

func (g *GithubMockServer) handleToken(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	code := r.FormValue("code")
	g.mu.Lock()
	username := g.oauthCodes[code]
	g.mu.Unlock()
	if username == "" {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	accessToken := "token-" + username
	g.AddUser(username, accessToken, nil)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"access_token":"%s","token_type":"bearer"}`, accessToken)
}

// --- Helpers ---

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "token ") {
		return strings.TrimPrefix(auth, "token ")
	}
	return ""
}
