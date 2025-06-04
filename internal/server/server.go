package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/clems4ever/lgtm/internal/protocol"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

type clientInfo struct {
	conn *websocket.Conn
	// if the githubUser variable is not set, it means the connection is established but
	// the client have not registered yet.
	githubUser string
	repos      map[string]struct{} // set of "owner/repo"
}

var (
	upgrader = websocket.Upgrader{}
)

type Server struct {
	ctx  context.Context
	done context.CancelFunc

	oauth2Config *oauth2.Config

	sessionStore *sessions.CookieStore
	httpClient   *http.Client
	pingInterval time.Duration

	approvalEngine *ApprovalEngine

	mu               sync.Mutex
	clientInfoByConn map[*websocket.Conn]*clientInfo
	clientsByRepo    map[string][]*clientInfo

	asyncRequestsMu sync.Mutex
	asyncRequests   map[string]*protocol.ResponseFuture
	wg              sync.WaitGroup
}

func NewServer(oauth2Config *oauth2.Config, pingInterval time.Duration) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		oauth2Config:     oauth2Config,
		approvalEngine:   NewApprovalEngine(),
		clientInfoByConn: make(map[*websocket.Conn]*clientInfo),
		clientsByRepo:    make(map[string][]*clientInfo),
		asyncRequests:    make(map[string]*protocol.ResponseFuture),
		ctx:              ctx,
		done:             cancel,
		pingInterval:     pingInterval,
	}
}

func (s *Server) Close() {
	s.done()

	s.asyncRequestsMu.Lock()
	for _, r := range s.asyncRequests {
		r.Close()
	}
	s.asyncRequestsMu.Unlock()
	s.wg.Wait()
}
