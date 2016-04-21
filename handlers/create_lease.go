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
)

type createLeaseReq struct {
	MaxTimeSec int `json:"max_time"`
}

type createLeaseResp struct {
	KubeConfig  string `json:"kubeconfig"`
	IP          string `json:"ip"`
	Token       string `json:"uuid"`
	ClusterName string `json:"cluster_name"`
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

		availableCluster, err := searchForFreeCluster(clusterMap, leaseMap)
		if err != nil {
			switch e := err.(type) {
			case errNoAvailableOrExpiredClustersFound:
				htp.Error(w, http.StatusConflict, "no available clusters found")
				return
			case errExpiredLeaseGKEMissing:
				htp.Error(w, http.StatusInternalServerError, "cluster %s has an expired lease but doesn't exist in GKE", e.clusterName)
				return
			default:
				htp.Error(w, http.StatusInternalServerError, "unknown error %s", e.Error())
				return
			}
		}

		newToken := uuid.NewUUID()
		kubeConfig, err := createKubeConfigFromCluster(availableCluster)
		if err != nil {
			htp.Error(
				w,
				http.StatusInternalServerError,
				"error creating kubeconfig file for cluster %s (%s)",
				availableCluster.Name,
				err,
			)
			return
		}
		kubeConfigStr, err := marshalAndEncodeKubeConfig(kubeConfig)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error marshaling & encoding kubeconfig (%s)", err)
			return
		}

		resp := createLeaseResp{
			KubeConfig:  kubeConfigStr,
			IP:          availableCluster.Endpoint,
			Token:       newToken.String(),
			ClusterName: availableCluster.Name,
		}

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
