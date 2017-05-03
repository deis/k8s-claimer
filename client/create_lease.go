package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/htp"
)

// CreateLease creates a lease
func CreateLease(server, authToken, cloudProvider, clusterVersion, clusterRegex string, durationSec int) (*api.CreateLeaseResp, error) {
	endpt := newEndpoint(htp.Post, server, "lease")
	reqBuf := new(bytes.Buffer)
	req := api.CreateLeaseReq{MaxTimeSec: durationSec, ClusterRegex: clusterRegex, ClusterVersion: clusterVersion, CloudProvider: cloudProvider}
	if err := json.NewEncoder(reqBuf).Encode(req); err != nil {
		return nil, errEncoding{err: err}
	}
	res, err := endpt.executeReq(getHTTPClient(), reqBuf, authToken)
	if err != nil {
		return nil, errHTTPRequest{endpoint: endpt.String(), err: err}
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		message := string(bodyBytes)
		return nil, APIError{endpoint: endpt.String(), code: res.StatusCode, message: message}
	}

	decodedRes, err := api.DecodeCreateLeaseResp(res.Body)
	if err != nil {
		return nil, errDecoding{err: err}
	}

	return decodedRes, nil
}
