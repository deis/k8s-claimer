package handlers

import (
	"net/http"
)

// DeleteLease returns the http handler for the DELETE /lease/{token} endpoint
func DeleteLease() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
