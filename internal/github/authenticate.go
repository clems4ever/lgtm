package github

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type AuthCodeReceiver interface {
	Receive(code string)
}

type AuthCodeWaiter interface {
	Wait(url string) (string, error)
}

// AuthCallback combines the functionality of receiving and waiting for an authentication code.
// It is used to facilitate the OAuth2 authentication flow.
type AuthCallback interface {
	AuthCodeReceiver // Interface for receiving an authentication code
	AuthCodeWaiter   // Interface for waiting for an authentication code
}

// Authenticate performs the OAuth2 authentication flow with GitHub.
// It first attempts to load an existing token from disk. If no token is found,
// it initiates the OAuth2 flow by opening a browser for the user to log in and retrieve an auth code.
// The retrieved token is then saved to disk for future use.
//
// Parameters:
// - ctx: Context for managing the lifecycle of the authentication process.
// - tokenPath: Path to the file where the token is stored.
// - oauthConfig: OAuth2 configuration for GitHub.
// - baseURL: Base URL for GitHub API requests.
// - waiter: Callback for handling the auth code retrieval process.
// - client: HTTP client for making requests.
//
// Returns:
// - A GitHub client initialized with the retrieved token.
// - An error if the authentication process fails.
func Authenticate(
	ctx context.Context,
	tokenPath string,
	oauthConfig *oauth2.Config,
	baseURL string,
	waiter AuthCodeWaiter,
	client *http.Client,
) (*Client, error) {
	// Attempt to load an existing token from disk
	token, err := LoadToken(tokenPath)
	if err == nil {
		fmt.Println("✅ Github token loaded from disk")
		return NewClient(token.AccessToken, baseURL, client), nil
	}

	// Generate the URL for the OAuth2 authorization flow
	url := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Wait for the user to provide the auth code via the callback
	code, err := waiter.Wait(url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve auth code: %w", err)
	}

	// Exchange the auth code for an access token
	token, err = oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth code: %w", err)
	}

	// Save the token to disk for future use
	err = SaveToken(tokenPath, token)
	if err != nil {
		return nil, fmt.Errorf("failed to save token on disk: %w", err)
	}
	fmt.Println("✅ Logged in and token saved to disk")

	// Return a new GitHub client initialized with the access token
	return NewClient(token.AccessToken, baseURL, client), nil
}
