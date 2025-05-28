package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRepos_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user/repos" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `[{"full_name":"foo/bar","permissions":{"push":true}},{"full_name":"baz/qux","permissions":{"push":false}}]`)
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

	repos, err := client.GetRepos()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0] != "foo/bar" {
		t.Errorf("expected [foo/bar], got %v", repos)
	}
}

func TestGetRepos_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer ts.Close()

	client := &Client{
		httpClient:  ts.Client(),
		accessToken: "dummy",
		apiBaseURL:  ts.URL,
	}

	_, err := client.GetRepos()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRepos_InvalidJSON(t *testing.T) {
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

	_, err := client.GetRepos()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRepos_NetworkError(t *testing.T) {
	client := &Client{
		httpClient:  &http.Client{},
		accessToken: "dummy",
		apiBaseURL:  "http://127.0.0.1:0", // invalid port, should fail to connect
	}

	_, err := client.GetRepos()
	if err == nil {
		t.Fatal("expected error due to network failure, got nil")
	}
}
