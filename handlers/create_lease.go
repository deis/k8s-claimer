package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/deis/k8s-claimer/htp"
	"github.com/pborman/uuid"
	container "google.golang.org/api/container/v1"
)

type createLeaseReq struct {
	MaxTimeSec int `json:"max_time"`
}

type createLeaseResp struct {
	KubeConfig string `json:"kubeconfig"`
	IP         string `json:"ip"`
	Token      string `json:"uuid"`
}

func (c createLeaseReq) maxTimeDur() time.Duration {
	return time.Duration(c.MaxTimeSec) * time.Second
}

// CreateLease creates the handler that responds to the POST /lease endpoint
func CreateLease(containerService *container.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := new(createLeaseReq)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			htp.Error(w, http.StatusBadRequest, "error decoding JSON (%s)", err)
			return
		}
		resp := createLeaseResp{KubeConfig: "", IP: "", Token: uuid.New()}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			htp.Error(w, http.StatusInternalServerError, "error encoding json (%s)", err)
			return
		}
	})
}
