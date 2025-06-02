package server

import "net/http"

// apiAuthMiddleware wraps an HTTP handler with authentication logic.
// It checks the Authorization header for a Bearer token and ensures it matches the provided authToken.
// If the token is invalid, the request is rejected with a 401 Unauthorized status.
func apiAuthMiddleware(authToken string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authToken == "" {
			fn(w, r)
			return
		}
		// Authenticate using the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+authToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		fn(w, r)
	}
}
