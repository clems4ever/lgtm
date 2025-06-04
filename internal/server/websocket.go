package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/clems4ever/lgtm/internal/github"
	"github.com/clems4ever/lgtm/internal/protocol"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	// ErrNoEligibleApprover is returned when no eligible approver is found for a PR.
	ErrNoEligibleApprover = fmt.Errorf("no eligible approver")
)

// wsHandler handles WebSocket connections for client registration and PR approval requests.
// It upgrades the HTTP connection to WebSocket, processes registration messages,
// and maintains the list of connected clients and their repositories.
func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Initialize client information for this connection
	var info clientInfo
	info.conn = conn
	info.repos = make(map[string]struct{})

	// Register the new client in the server's state
	s.mu.Lock()
	s.clientInfoByConn[conn] = &info
	s.mu.Unlock()

	log.Printf("new client connected!")

	// Listen for messages from the client
	for {
		var message protocol.Message
		err := protocol.Read(conn, &message)
		if err != nil {
			break
		}

		// Handle async responses or registration messages
		s.asyncRequestsMu.Lock()
		req, ok := s.asyncRequests[message.RequestID]
		if ok {
			s.handleAsyncResponse(message.RequestID, req, message.Message)
			s.asyncRequestsMu.Unlock()
			continue
		}
		s.asyncRequestsMu.Unlock()
		s.handleAsyncMessage(message.Message, &info)
	}

	// Clean up the client on disconnection
	s.mu.Lock()
	delete(s.clientInfoByConn, conn)
	// Remove the client from all repositories it was registered for
	for repo := range info.repos {
		list := s.clientsByRepo[repo]
		newList := make([]*clientInfo, 0, len(list))
		for _, c := range list {
			if c != &info {
				newList = append(newList, c)
			}
		}
		if len(newList) == 0 {
			delete(s.clientsByRepo, repo)
		} else {
			s.clientsByRepo[repo] = newList
		}
	}
	s.mu.Unlock()

	s.approvalEngine.RemoveApprover(info.githubUser)
	if info.githubUser != "" {
		log.Printf("client disconnected (user %s)\n", info.githubUser)
	} else {
		log.Printf("anonymous client disconnected\n")
	}
}

// handleAsyncMessage processes protocol messages that are not responses to async requests.
func (s *Server) handleAsyncMessage(msg any, info *clientInfo) {
	switch v := msg.(type) {
	case protocol.RegisterRequestMessage:
		err := s.handleRegisterRequestMessage(v, info)
		if err != nil {
			log.Printf("failed to handle message: %s", err)
		}
	case protocol.PingMessage:
		// fmt.Println("ping")
	default:
		log.Println("unsupported message type")
	}
}

// handleAsyncResponse delivers an async response to the waiting future and cleans up the request.
func (s *Server) handleAsyncResponse(requestID string, future *protocol.ResponseFuture, msg any) {
	future.ReceiveResponse(msg)
	s.cleanupAsyncRequest(requestID)
}

// RequestApproval forwards a pull request approval request to an eligible approver.
// It selects a random approver from the list of connected clients who are registered for the target repository.
// If no eligible approver is found, returns ErrNoEligibleApprover.
func (s *Server) RequestApproval(link github.PRLink) error {
	fmt.Println("need to forward approval link:", link)
	targetRepo := link.RepoFullName()

	s.mu.Lock()
	eligible := []*clientInfo{}
	for _, c := range s.clientsByRepo[targetRepo] {
		eligible = append(eligible, c)
	}
	s.mu.Unlock()

	// TODO: rewrite this without recursion.
	return s.routePRApprovalRequestRecursive(link, eligible)
}

// routePRApprovalRequestRecursive tries to forward the approval request to eligible clients, recursively excluding authors.
func (s *Server) routePRApprovalRequestRecursive(link github.PRLink, eligible []*clientInfo) error {
	if len(eligible) == 0 {
		fmt.Printf("no eligible approver for %s\n", link)
		return ErrNoEligibleApprover
	}

	// Select a random eligible client
	idx := 0
	if len(eligible) > 1 {
		idx = int(time.Now().UnixNano()) % len(eligible)
	}
	selected := eligible[idx]

	fmt.Printf("%s will tentatively be approved by %s\n", link, selected.githubUser)

	res, _, err := s.sendRPC(selected.conn, protocol.ApproveRequestMessage{
		Link: link,
	}, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to send rpc call: %w", err)
	}

	// Wait for the approval response
	var resp protocol.ApproveResponseMessage
	err = res.WaitResponse(&resp)
	if err != nil {
		return fmt.Errorf("failed to receive async response: %w", err)
	}

	switch resp.Response {
	case protocol.ApproveResponseSuccess:
		fmt.Printf("%s approved by %s\n", link, selected.githubUser)
		return nil
	case protocol.ApproveResponseErrSameAuthor:
		fmt.Printf("%s not approved by author %s\n", link, selected.githubUser)
		// Create a reduced list excluding the author
		reducedList := make([]*clientInfo, 0, len(eligible))
		for _, c := range eligible {
			if c.githubUser == selected.githubUser {
				continue
			}
			reducedList = append(reducedList, c)
		}
		return s.routePRApprovalRequestRecursive(link, reducedList)
	}
	return fmt.Errorf("%s", resp.Response)
}

// sendRPC sends an approval request message to a client and sets up a ResponseFuture for the response.
func (s *Server) sendRPC(conn *websocket.Conn, msg protocol.ApproveRequestMessage, timeout time.Duration) (*protocol.ResponseFuture, string, error) {
	requestID := uuid.NewString()
	err := protocol.WriteWithRequestID(conn, msg, requestID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to write request with id: %w", err)
	}

	res := protocol.NewResponseFuture()
	s.asyncRequestsMu.Lock()
	s.asyncRequests[requestID] = res
	s.asyncRequestsMu.Unlock()
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		select {
		case <-s.ctx.Done():
		case <-time.After(timeout):
			res.Close()
		}
		s.asyncRequestsMu.Lock()
		s.cleanupAsyncRequest(requestID)
		s.asyncRequestsMu.Unlock()
	}()
	return res, requestID, nil
}

// cleanupAsyncRequest removes a request from the asyncRequests map.
func (s *Server) cleanupAsyncRequest(reqID string) {
	if _, ok := s.asyncRequests[reqID]; ok {
		delete(s.asyncRequests, reqID)
	}
}

// handleRegisterRequestMessage processes a registration message from a client and updates the server state.
func (s *Server) handleRegisterRequestMessage(msg protocol.RegisterRequestMessage, info *clientInfo) error {
	log.Printf("client registered with handle %s\n", msg.GithubUser)
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store the client's repositories and GitHub user
	for _, repo := range msg.Repos {
		info.repos[repo] = struct{}{}
	}
	info.githubUser = msg.GithubUser

	// Update global state with the new client
	for _, repo := range msg.Repos {
		s.clientsByRepo[repo] = append(s.clientsByRepo[repo], info)
	}

	s.approvalEngine.AddApprover(msg.GithubUser)
	return nil
}
