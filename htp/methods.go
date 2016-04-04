package htp

import (
	"net/http"
)

const (
	// Get is the constant for the HTTP GET method
	Get Method = "GET"
	// Post is the constant for the HTTP POST method
	Post Method = "POST"
	// Put is the constant for the HTTP PUT method
	Put Method = "PUT"
	// Delete is the constant for the HTTP DELETE method
	Delete Method = "DELETE"
	// Head is the constant for the HTTP HEAD method
	Head Method = "HEAD"
	// Options is the constant for the HTTP OPTIONS method
	Options Method = "OPTIONS"
)

// Method is the type for the incoming HTTP Method
type Method string

// String is the fmt.Stringer interface implementation
func (m Method) String() string {
	return string(m)
}

// MatchesMethod returns whether or not the method in r is the same as m
func MatchesMethod(r *http.Request, m Method) bool {
	return r.Method == m.String()
}
