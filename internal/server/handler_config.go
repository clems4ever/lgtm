package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.cfg); err != nil {
		http.Error(w, "Failed to encode config", http.StatusInternalServerError)
	}
}
