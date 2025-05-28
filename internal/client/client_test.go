package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/clems4ever/lgtm/internal/client"
	"github.com/clems4ever/lgtm/internal/common"
	"github.com/clems4ever/lgtm/internal/test"
	"github.com/stretchr/testify/require"
)

// mockWsServer simulates a websocket server
func mockWebsocketServer(t *testing.T, connCh chan struct{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/config":
			err := json.NewEncoder(w).Encode(common.SharedConfig{
				ClientID:     "test",
				ClientSecret: "test",
			})
			require.NoError(t, err)
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
	return wsServer, githubSrv, tokenPath
}

func newClient(t *testing.T, ctx context.Context, port int) *client.Client {
	connCh := make(chan struct{})
	wsServer, githubSrv, tokenPath := newMockServers(t, port, connCh)

	authCallback := test.NewTestAuthCallback()

	// Prepare the client
	c := client.NewClient(wsServer.URL, "", fmt.Sprintf(":%d", port), tokenPath,
		githubSrv.URL(), githubSrv.URL(), authCallback, http.DefaultClient)

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

	err := c.Shutdown(ctx)
	require.NoError(t, err)
}

func TestClient_GetHomePage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	port := 8051
	c := newClient(t, ctx, port)

	res, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
	require.NoError(t, err)

	require.Equal(t, 200, res.StatusCode)

	err = c.Shutdown(ctx)
	require.NoError(t, err)
}
