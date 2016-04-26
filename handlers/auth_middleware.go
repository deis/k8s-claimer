package handlers

import "net/http"

// WithAuth provides handling of endpoints requiring authentication
func WithAuth(tokenToCheck, tokenHeaderName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(tokenHeaderName) != tokenToCheck {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
