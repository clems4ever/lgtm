// Tests for GetAuthenticatedUserLogin method of the GitHub client.
// These tests cover successful authentication, HTTP errors, invalid JSON responses, and network failures.
package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetAuthenticatedUserLogin_Success verifies that GetAuthenticatedUserLogin returns the correct username on a successful response.
func TestGetAuthenticatedUserLogin_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"login":"testuser"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := &Client{
		httpClient:  ts.Client(),
		accessToken: "dummy",
		apiBaseURL:  ts.URL,
	}

	login, err := client.GetAuthenticatedUserLogin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if login != "testuser" {
		t.Errorf("expected login 'testuser', got %q", login)
	}
}

// TestGetAuthenticatedUserLogin_HTTPError checks that an HTTP error from the server results in an error from GetAuthenticatedUserLogin.
func TestGetAuthenticatedUserLogin_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer ts.Close()

	client := &Client{
		httpClient:  ts.Client(),
		accessToken: "dummy",
		apiBaseURL:  ts.URL,
	}

	_, err := client.GetAuthenticatedUserLogin()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestGetAuthenticatedUserLogin_InvalidJSON ensures that invalid JSON in the response causes an error.
func TestGetAuthenticatedUserLogin_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `not a json`)
	}))
	defer ts.Close()

	client := &Client{
		httpClient:  ts.Client(),
		accessToken: "dummy",
		apiBaseURL:  ts.URL,
	}

	_, err := client.GetAuthenticatedUserLogin()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestGetAuthenticatedUserLogin_NetworkError verifies that a network failure results in an error from GetAuthenticatedUserLogin.
func TestGetAuthenticatedUserLogin_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate a network error
	client := &Client{
		httpClient:  &http.Client{Timeout: 1 * time.Second},
		accessToken: "dummy",
		apiBaseURL:  "http://127.0.0.1:0", // invalid port, should fail to connect
	}

	_, err := client.GetAuthenticatedUserLogin()
	if err == nil {
		t.Fatal("expected error due to network failure, got nil")
	}
}
