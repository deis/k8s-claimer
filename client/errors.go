package client

import (
	"fmt"
)

type errInvalidStatusCode struct {
	endpoint string
	code     int
}

func (e errInvalidStatusCode) Error() string {
	return fmt.Sprintf("invalid status code for endpoing (%s): %d", e.endpoint, e.code)
}

type errHTTPRequest struct {
	endpoint string
	err      error
}

func (e errHTTPRequest) Error() string {
	return fmt.Sprintf("error executing HTTP request on %s (%s)", e.endpoint, e.err)
}

type errEncoding struct {
	err error
}

func (e errEncoding) Error() string {
	return fmt.Sprintf("Error encoding request body (%s)", e.err)
}

type errDecoding struct {
	err error
}

func (e errDecoding) Error() string {
	return fmt.Sprintf("Error decoding response body (%s)", e.err)
}
