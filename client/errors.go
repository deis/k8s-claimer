package client

import (
	"fmt"
)

// APIError is an error returned from the api
type APIError struct {
	endpoint string
	code     int
	message  string
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

type errHTTPRequest struct {
	endpoint string
	err      error
}

func (e errHTTPRequest) Error() string {
	return fmt.Sprintf("Error executing HTTP request on %s -- %s", e.endpoint, e.err)
}

type errEncoding struct {
	err error
}

func (e errEncoding) Error() string {
	return fmt.Sprintf("Error encoding request body -- %s", e.err)
}

type errDecoding struct {
	err error
}

func (e errDecoding) Error() string {
	return fmt.Sprintf("Error decoding response body -- %s", e.err)
}
