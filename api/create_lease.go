package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"time"
)

// CreateLeaseReq is the encoding/json compatible struct that represents the POST /lease
// request body
type CreateLeaseReq struct {
	MaxTimeSec int `json:"max_time"`
}

// MaxTimeDur returns the maximum time specified in c as a time.Duration
func (c CreateLeaseReq) MaxTimeDur() time.Duration {
	return time.Duration(c.MaxTimeSec) * time.Second
}

// ExpirationTime is a convenience function for start.Add(c.MaxTimeDur())
func (c CreateLeaseReq) ExpirationTime(start time.Time) time.Time {
	return start.Add(c.MaxTimeDur())
}

// CreateLeaseResp is the encoding/json compatible struct that represents the POST /lease
// response body
type CreateLeaseResp struct {
	KubeConfig  string `json:"kubeconfig"`
	IP          string `json:"ip"`
	Token       string `json:"uuid"`
	ClusterName string `json:"cluster_name"`
}

// DecodeCreateLeaseResp decodes rdr from its JSON representation into a CreateLeaseResp.
// If there was any error reading rdr or it had malformed JSON, returns nil and the error
func DecodeCreateLeaseResp(rdr io.Reader) (*CreateLeaseResp, error) {
	ret := new(CreateLeaseResp)
	if err := json.NewDecoder(rdr).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// DecodeKubeConfig decodes c.KubeConfig by the RFC 4648 standard.
// See http://tools.ietf.org/html/rfc4648 for more information
func (c CreateLeaseResp) DecodeKubeConfig() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.KubeConfig)
}
