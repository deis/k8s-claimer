package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/htp"
)

// DeleteLease deletes a lease
func DeleteLease(server, authToken, cloudProvider, leaseToken string) error {
	endpt := newEndpoint(htp.Delete, server, "lease/"+leaseToken)
	reqBuf := new(bytes.Buffer)
	req := api.DeleteLeaseReq{CloudProvider: cloudProvider}
	if err := json.NewEncoder(reqBuf).Encode(req); err != nil {
		return nil, errEncoding{err: err}
	}
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
