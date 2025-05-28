package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/clems4ever/lgtm/internal/protocol"
)

// wsHandler handles WebSocket connections for client registration.
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

	// Initialize client information
	var info clientInfo
	info.conn = conn
	info.repos = make(map[string]struct{})

	// Update global state with the new client
	mu.Lock()
	clients[conn] = &info
	mu.Unlock()

	log.Printf("new client connected!")

	// Keep the WebSocket connection open and listen for messages
	for {
		var message protocol.Message
		err := protocol.Read(conn, &message)
		if err != nil {
			break
		}

		// Handle different types of messages
		switch v := message.Message.(type) {
		case protocol.RegisterRequestMessage:
			err = handleRegisterRequestMessage(v, &info)
			if err != nil {
				log.Printf("failed to handle message: %s", err)
			}
		case protocol.ApproveMessage:
			err = handleApprovalMessage(v)
			if err != nil {
				log.Printf("failed to handle message: %s", err)
			}
		default:
			log.Println("unsupported message type")
		}
	}

	// Clean up the client on disconnection
	mu.Lock()
	delete(clients, conn)
	// Remove the client from all repositories it was registered for
	for repo := range info.repos {
		list := clientsByRepo[repo]
		newList := make([]*clientInfo, 0, len(list))
		for _, c := range list {
			if c != &info {
				newList = append(newList, c)
			}
		}
		if len(newList) == 0 {
			delete(clientsByRepo, repo)
		} else {
			clientsByRepo[repo] = newList
		}
	}
	mu.Unlock()
	if info.githubUser != "" {
		log.Printf("client disconnected (user %s)\n", info.githubUser)
	} else {
		log.Printf("anonymous client disconnected\n")
	}
}

// handleApprovalMessage processes an ApproveMessage and forwards it to an eligible client.
// It selects a random eligible client (excluding the submitter) and sends the approval request.
func handleApprovalMessage(msg protocol.ApproveMessage) error {
	fmt.Println("need to forward approval link: ", msg.Link)
	targetRepo := msg.Link.RepoFullName()

	mu.Lock()
	eligible := []*clientInfo{}
	for _, c := range clientsByRepo[targetRepo] {
		// Exclude the submitter's GitHub user
		if c.githubUser != msg.Author {
			eligible = append(eligible, c)
		}
	}
	mu.Unlock()

	if len(eligible) == 0 {
		log.Printf("No eligible approver for repo %s (excluding submitter)", targetRepo)
		return nil
	}

	// Select a random eligible client
	idx := 0
	if len(eligible) > 1 {
		idx = int(time.Now().UnixNano()) % len(eligible)
	}
	selected := eligible[idx]

	fmt.Printf("PR %s will be approved by %s\n", msg.Link, selected.githubUser)

	// Forward the approval message to the selected client
	forwardMsg := protocol.Message{
		Type:    protocol.ApproveMessageType,
		Message: msg,
	}
	if err := selected.conn.WriteJSON(forwardMsg); err != nil {
		log.Printf("Failed to forward approval message to client: %v", err)
	}
	return nil
}

func handleRegisterRequestMessage(msg protocol.RegisterRequestMessage, info *clientInfo) error {
	log.Printf("client registered with handle %s\n", msg.GithubUser)
	mu.Lock()
	// Store the client's repositories and GitHub user
	for _, repo := range msg.Repos {
		info.repos[repo] = struct{}{}
	}
	info.githubUser = msg.GithubUser

	// Update global state with the new client
	for _, repo := range msg.Repos {
		clientsByRepo[repo] = append(clientsByRepo[repo], info)
	}
	users[msg.GithubUser] = struct{}{}

	// Prepare the list of approvers for all clients
	approvers := []string{}
	for user := range users {
		approvers = append(approvers, user)
	}
	mu.Unlock()
	// Notify all clients about the updated list of approvers
	for sock := range clients {
		err := protocol.Write(sock, protocol.ApproverEventMessage{
			Approvers: approvers,
		})
		if err != nil {
			log.Printf("failed to send approver event: %s\n", err)
		}
	}
	return nil
}
