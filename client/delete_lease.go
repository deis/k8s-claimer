package client

import (
	"net/http"

	"github.com/deis/k8s-claimer/htp"
)

// DeleteLease deletes a lease
func DeleteLease(server, authToken, leaseToken string) error {
	endpt := newEndpoint(htp.Delete, server, "lease/"+leaseToken)
	resp, err := endpt.executeReq(getHTTPClient(), nil, authToken)
	if err != nil {
		return errHTTPRequest{endpoint: endpt.String(), err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errInvalidStatusCode{endpoint: endpt.String(), code: resp.StatusCode}
	}
	return nil
}
