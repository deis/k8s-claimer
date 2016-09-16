package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/deis/k8s-claimer/htp"
)

type endpoint struct {
	host   string
	path   string
	method htp.Method
}

func newEndpoint(method htp.Method, host, path string) endpoint {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return endpoint{host: host, path: path, method: method}
}

func (e endpoint) String() string {
	return fmt.Sprintf("%s %s/%s", e.method, e.host, e.path)
}

func (e endpoint) httpReq(body io.Reader) (*http.Request, error) {
	return http.NewRequest(e.method.String(), fmt.Sprintf("%s/%s", e.host, e.path), body)
}

func (e endpoint) executeReq(cl *http.Client, body io.Reader, authToken string) (*http.Response, error) {
	req, err := e.httpReq(body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authToken)
	return cl.Do(req)
}

func getHTTPClient() *http.Client {
	return http.DefaultClient
}
