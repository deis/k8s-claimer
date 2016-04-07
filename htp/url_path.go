package htp

import (
	"net/http"
	"strings"
)

// SplitPath returns a slice of path elements in r. It returns nil if r.URL.Path is empty or "/"
func SplitPath(r *http.Request) []string {
	if len(r.URL.Path) == 0 {
		return nil
	}
	if len(r.URL.Path) == 1 && r.URL.Path[0] == '/' {
		return nil
	}
	return strings.Split(r.URL.Path[1:], "/")
}
