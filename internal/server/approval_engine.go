package server

import (
	"sync"
)

type ApprovalEngine struct {
	mu                 sync.RWMutex
	availableApprovers map[string]int
}

func NewApprovalEngine() *ApprovalEngine {
	return &ApprovalEngine{
		availableApprovers: make(map[string]int),
	}
}

func (ae *ApprovalEngine) GetApprovers() []string {
	var l []string
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	for a := range ae.availableApprovers {
		l = append(l, a)
	}
	return l
}

func (ae *ApprovalEngine) AddApprover(user string) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	count, ok := ae.availableApprovers[user]
	if !ok {
		ae.availableApprovers[user] = 1
	} else {
		ae.availableApprovers[user] = count + 1
	}
}

func (ae *ApprovalEngine) RemoveApprover(user string) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	count, ok := ae.availableApprovers[user]
	if !ok {
		return
	} else if count == 1 {
		delete(ae.availableApprovers, user)
	} else {
		ae.availableApprovers[user] = count - 1
	}
}
