package github

import (
	"net/http"
	"testing"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("token123", "", nil)
	if c.accessToken != "token123" {
		t.Errorf("expected accessToken to be 'token123', got %q", c.accessToken)
	}
	if c.apiBaseURL != "https://api.github.com" {
		t.Errorf("expected apiBaseURL to be 'https://api.github.com', got %q", c.apiBaseURL)
	}
	if c.httpClient != http.DefaultClient {
		t.Errorf("expected httpClient to be http.DefaultClient")
	}
}

func TestNewClientCustom(t *testing.T) {
	customClient := &http.Client{}
	c := NewClient("abc", "https://custom.api", customClient)
	if c.accessToken != "abc" {
		t.Errorf("expected accessToken to be 'abc', got %q", c.accessToken)
	}
	if c.apiBaseURL != "https://custom.api" {
		t.Errorf("expected apiBaseURL to be 'https://custom.api', got %q", c.apiBaseURL)
	}
	if c.httpClient != customClient {
		t.Errorf("expected httpClient to be customClient")
	}
}
