package htp

import (
	"net/http"
)

// MethodMux returns an http.Handler that routes to the correct handler based on the request
// method as defined in m
func MethodMux(m map[Method]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := m[Method(r.Method)]
		if !ok {
			Error(w, http.StatusNotFound, "%s %s not found", r.Method, r.URL.Path)
			return
		}
		h.ServeHTTP(w, r)
	})
}
