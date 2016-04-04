package htp

import (
	"fmt"
	"net/http"
)

// Error is a convenience function for http.Error, but it allows you to specify the error string
// in a format string
func Error(w http.ResponseWriter, code int, fmtStr string, vars ...interface{}) {
	http.Error(w, fmt.Sprintf(fmtStr, vars...), code)
}
