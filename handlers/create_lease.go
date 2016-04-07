package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
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

func (c createLeaseReq) expirationTime(start time.Time) time.Time {
	return start.Add(c.maxTimeDur())
}

// CreateLease creates the handler that responds to the POST /lease endpoint
func CreateLease(
	clusterLister gke.ClusterLister,
	services k8s.ServiceGetterUpdater,
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

		clusterMap, err := clusters.ParseMapFromGKE(clusterLister, gCloudProjID, gCloudZone)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error fetching GKE clusters (%s)", err)
			return
		}

		leaseMap, err := leases.ParseMapFromAnnotations(svc.Annotations)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error parsing leases from Kubernetes annotations (%s)", err)
			return
		}

		unusedCluster, unusedClusterErr := findUnusedGKECluster(clusterMap, leaseMap)
		uuidAndLease, expiredLeaseErr := findExpiredLease(leaseMap)
		if unusedClusterErr != nil && expiredLeaseErr != nil {
			htp.Error(w, http.StatusConflict, "no available or expired clusters found")
			return
		}
		var availableCluster *container.Cluster
		if unusedCluster == nil {
			availableCluster = unusedCluster
		}
		if expiredLeaseErr == nil {
			cl, ok := clusterMap.ClusterByName(uuidAndLease.Lease.ClusterName)
			if !ok {
				htp.Error(w, http.StatusInternalServerError, "cluster %s has an expired lease but does not exist in GKE", uuidAndLease.Lease.ClusterName)
				return
			}
			availableCluster = cl
		}

		newToken := uuid.NewUUID()
		resp := createLeaseResp{
			KubeConfig: createKubeConfigFromCluster(availableCluster),
			IP:         availableCluster.Endpoint,
			Token:      newToken.String(),
		}
		leaseMap.DeleteLease(uuidAndLease.UUID)
		leaseMap.CreateLease(newToken, leases.NewLease(availableCluster.Name, req.expirationTime(time.Now())))

		if err := saveAnnotations(services, svc, leaseMap); err != nil {
			htp.Error(w, http.StatusInternalServerError, "error saving new lease to Kubernetes annotations (%s)", err)
			return
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			htp.Error(w, http.StatusInternalServerError, "error encoding json (%s)", err)
			return
		}
	})
}
