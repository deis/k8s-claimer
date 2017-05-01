package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	"github.com/deis/k8s-claimer/k8s"
)

// CreateLeaseReq is the encoding/json compatible struct that represents the POST /lease
// request body
type CreateLeaseReq struct {
	MaxTimeSec     int    `json:"max_time"`
	ClusterRegex   string `json:"cluster_regex"`
	ClusterVersion string `json:"cluster_version"`
	CloudProvider  string `json:"cloud_provider"`
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
	KubeConfigStr  string `json:"kubeconfig"`
	IP             string `json:"ip"`
	Token          string `json:"uuid"`
	ClusterName    string `json:"cluster_name"`
	ClusterVersion string `json:"cluster_version"`
	CloudProvider  string `json:"cloud_provider"`
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

// KubeConfigBytes decodes c.KubeConfig by the RFC 4648 standard.
// See http://tools.ietf.org/html/rfc4648 for more information
func (c CreateLeaseResp) KubeConfigBytes() ([]byte, error) {
	kubeConfigBytes, err := base64.StdEncoding.DecodeString(c.KubeConfigStr)
	if err != nil {
		return nil, err
	}
	return kubeConfigBytes, nil
}

// KubeConfig returns decoded and unmarshalled Kubernetes client configuration
func (c CreateLeaseResp) KubeConfig() (*k8s.KubeConfig, error) {
	configBytes, err := c.KubeConfigBytes()
	if err != nil {
		return nil, err
	}
	kubeConfig := &k8s.KubeConfig{}
	if err := json.Unmarshal(configBytes, kubeConfig); err != nil {
		return nil, err
	}
	return kubeConfig, nil
}
