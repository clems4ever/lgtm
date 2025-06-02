package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/clems4ever/lgtm/internal/client"
	"github.com/clems4ever/lgtm/internal/test"
	"github.com/stretchr/testify/require"
)

// mockWsServer simulates a websocket server
func mockWebsocketServer(t *testing.T, connCh chan struct{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ws":
			connCh <- struct{}{}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"login":"mockuser"}`)
		default:
			http.NotFound(w, r)
		}
	}))
}

// mockServer is a minimal HTTP server for the local client server.
type mockServer struct {
	started chan struct{}
	stopped chan struct{}
}

func (m *mockServer) Start() {
	close(m.started)
	<-m.stopped
}

func newMockServers(t *testing.T, port int, connCh chan struct{}) (*httptest.Server, *test.GithubMockServer, string) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	wsServer := mockWebsocketServer(t, connCh)
	t.Cleanup(func() {
		wsServer.Close()
	})

	// Set up mock OAuth2 and GitHub servers
	githubSrv := test.NewGithubMockServer(t, fmt.Sprintf("http://localhost:%d/callback", port))
	t.Cleanup(func() {
		githubSrv.Close()
	})

	githubSrv.AddUser("testuser", "access-token", []test.Repo{
		{
			FullName: "testuser/myrepo",
			Permissions: test.RepoPermissions{
				Push: true,
			},
		},
	})
	return wsServer, githubSrv, tokenPath
}

func newClient(t *testing.T, ctx context.Context, port int) *client.Client {
	connCh := make(chan struct{})
	wsServer, githubSrv, _ := newMockServers(t, port, connCh)

	// Prepare the client
	c, err := client.NewClient(wsServer.URL, "", time.Second, "access-token",
		githubSrv.URL(), http.DefaultClient)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := c.Start()
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		wg.Wait()
	})

	select {
	case <-connCh:
	case <-ctx.Done():
		require.FailNow(t, ctx.Err().Error())
	}
	return c
}

func TestClient_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := newClient(t, ctx, 8050)

	c.Stop()
}
