package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/deis/k8s-claimer/htp"
	"github.com/pborman/uuid"
	container "google.golang.org/api/container/v1"
	"k8s.io/kubernetes/pkg/api"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
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
func CreateLease(
	containerService *container.Service,
	services k8s.ServiceInterface,
	k8sServiceName,
	gCloudProjID,
	gCloudZone string,
) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := new(createLeaseReq)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			htp.Error(w, http.StatusBadRequest, "error decoding JSON (%s)", err)
			return
		}

		svc, err := services.Get(k8sServiceName)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error listing service (%s)", err)
			return
		}

		gkeClusters, err := getGKEClusters(containerService, projID, zone)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error listing GKE clusters (%s)", err)
		}

		unusedCluster, err := findUnusedGKECluster(gkeClusters, svc)
		if err != nil {
			htp.Error(w, http.StatusConflict, "un-leased cluster not found")
		}

		// findExpiredLease(...)

		resp := createLeaseResp{KubeConfig: "", IP: "", Token: uuid.New()}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			htp.Error(w, http.StatusInternalServerError, "error encoding json (%s)", err)
			return
		}
	})
}
