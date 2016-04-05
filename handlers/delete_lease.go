package handlers

import (
	"net/http"

	k8s "k8s.io/kubernetes/pkg/client/unversioned"
)

// DeleteLease returns the http handler for the DELETE /lease/{token} endpoint
func DeleteLease(services k8s.ServiceInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
