package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/deis/k8s-claimer/htp"
)

// DeleteLeaseReq is the encoding/json compatible struct that represents the DELETE /lease request body
type DeleteLeaseReq struct {
	CloudProvider string `json:"cloud_provider"`
}

// DeleteLease deletes a lease
func DeleteLease(server, authToken, cloudProvider, leaseToken string) error {
	endpt := newEndpoint(htp.Delete, server, "lease/"+leaseToken)
	reqBuf := new(bytes.Buffer)
	req := DeleteLeaseReq{CloudProvider: cloudProvider}
	if err := json.NewEncoder(reqBuf).Encode(req); err != nil {
		return errEncoding{err: err}
	}
	resp, err := endpt.executeReq(getHTTPClient(), nil, authToken)
	if err != nil {
		return errHTTPRequest{endpoint: endpt.String(), err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return APIError{endpoint: endpt.String(), code: resp.StatusCode}
	}
	return nil
}
