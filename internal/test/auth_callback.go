package test

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// TestAuthCallback simulates an OAuth2 callback handler for tests.
type TestAuthCallback struct {
	codeCh chan string
}

// NewTestAuthCallback creates a new TestAuthCallback and stores the auth URL.
func NewTestAuthCallback() *TestAuthCallback {
	return &TestAuthCallback{
		codeCh: make(chan string),
	}
}

// Receive simulates receiving an auth code (not used in this mock).
func (c *TestAuthCallback) Receive(code string) {
	c.codeCh <- code
}

// Wait calls the /auth endpoint of the configured URL and waits for the code in the redirect.
func (c *TestAuthCallback) Wait(url string) (string, error) {
	go func() {
		// Call the /auth endpoint and follow the redirect to extract the code.
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to call /auth: %w", err))
		}
		defer resp.Body.Close()

		// Expect a redirect to /callback?code=...
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatal(fmt.Errorf("unexpected status from /auth: %d, body: %s", resp.StatusCode, string(body)))
		}

	}()
	return <-c.codeCh, nil
}
