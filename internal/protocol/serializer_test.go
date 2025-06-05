package protocol

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clems4ever/lgtm/internal/github"
	"github.com/gorilla/websocket"
)

// mockConn is a minimal wrapper to test Read/Write using a real websocket connection.
type mockConn struct {
	server     *httptest.Server
	client     *websocket.Conn
	serverConn *websocket.Conn
}

func newMockConn(t *testing.T) *mockConn {
	upgrader := websocket.Upgrader{}
	ch := make(chan *websocket.Conn, 1)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Upgrade error: %v", err)
		}
		ch <- c
	}))
	t.Cleanup(s.Close)

	u := "ws" + s.URL[len("http"):] // ws://127.0.0.1:xxxx/
	client, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	serverConn := <-ch
	return &mockConn{server: s, client: client, serverConn: serverConn}
}

func (mc *mockConn) close() {
	mc.client.Close()
	mc.serverConn.Close()
	mc.server.Close()
}

func TestWriteAndReadApproveMessage(t *testing.T) {
	mc := newMockConn(t)
	defer mc.close()

	orig := ApproveRequestMessage{
		Link: github.PRLink{Owner: "foo", Repo: "bar", PRNumber: 42},
	}
	go func() {
		_, err := Write(mc.serverConn, orig)
		if err != nil {
			t.Errorf("Write error: %v", err)
		}
	}()

	var msg Message
	err := Read(mc.client, &msg)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if msg.Type != ApproveRequestMessageType {
		t.Errorf("expected Type Approve, got %v", msg.Type)
	}
	if msg.RequestID == "" {
		t.Errorf("expected non-empty requestID")
	}
	got, ok := msg.Message.(ApproveRequestMessage)
	if !ok {
		t.Fatalf("expected ApproveMessage, got %T", msg.Message)
	}
	if got != orig {
		t.Errorf("ApproveMessage mismatch: got %+v, want %+v", got, orig)
	}
}

func TestWriteAndReadRegisterMessage(t *testing.T) {
	mc := newMockConn(t)
	defer mc.close()

	orig := RegisterRequestMessage{
		Repos:      []string{"foo/bar"},
		GithubUser: "octocat",
	}
	go func() {
		_, err := Write(mc.serverConn, orig)
		if err != nil {
			t.Errorf("Write error: %v", err)
		}
	}()

	var msg Message
	err := Read(mc.client, &msg)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if msg.Type != RegisterRequestMessageType {
		t.Errorf("expected Type Register, got %v", msg.Type)
	}
	if msg.RequestID == "" {
		t.Errorf("expected non-empty requestID")
	}
	got, ok := msg.Message.(RegisterRequestMessage)
	if !ok {
		t.Fatalf("expected RegisterMessage, got %T", msg.Message)
	}
	if got.GithubUser != orig.GithubUser || len(got.Repos) != len(orig.Repos) || got.Repos[0] != orig.Repos[0] {
		t.Errorf("RegisterMessage mismatch: got %+v, want %+v", got, orig)
	}
}

func TestWriteUnsupportedType(t *testing.T) {
	mc := newMockConn(t)
	defer mc.close()
	_, err := Write(mc.client, struct{ Foo string }{Foo: "bar"})
	if err == nil {
		t.Error("expected error for unsupported msg type, got nil")
	}
}
