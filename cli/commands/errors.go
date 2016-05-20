package commands

import (
	"errors"
	"fmt"
)

var (
	errMissingAuthToken      = errors.New("missing AUTH_TOKEN")
	errMissingLeaseToken     = errors.New("missing lease token")
	errMissingServer         = errors.New("missing IP")
	errMissingKubeConfigFile = errors.New("missing kubeconfig file")
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
