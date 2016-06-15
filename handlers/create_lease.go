package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"github.com/pborman/uuid"
	k8sapi "k8s.io/kubernetes/pkg/api"
)

func getSvcsAndClusters(
	clusterLister gke.ClusterLister,
	services k8s.ServiceGetterUpdater,
	gCloudProjID,
	gCloudZone,
	k8sServiceName string,
) (*clusters.Map, *k8sapi.Service, error) {

	errCh := make(chan error)
	doneCh := make(chan struct{})
	clusterMapCh := make(chan *clusters.Map)
	apiServiceCh := make(chan *k8sapi.Service)
	defer close(doneCh)
	go func() {
		svc, err := services.Get(k8sServiceName)
		if err != nil {
			select {
			case errCh <- err:
			case <-doneCh:
			}
			return
		}
		select {
		case apiServiceCh <- svc:
		case <-doneCh:
		}
	}()
	go func() {
		clusterMap, err := clusters.ParseMapFromGKE(clusterLister, gCloudProjID, gCloudZone)
		if err != nil {
			select {
			case errCh <- err:
			case <-doneCh:
			}
			return
		}
		select {
		case clusterMapCh <- clusterMap:
		case <-doneCh:
		}
	}()

	var clusterMapRet *clusters.Map
	var apiServiceRet *k8sapi.Service
	for {
		select {
		case err := <-errCh:
			return nil, nil, err
		case cm := <-clusterMapCh:
			clusterMapRet = cm
		case svc := <-apiServiceCh:
			apiServiceRet = svc
		}
		if clusterMapRet != nil && apiServiceRet != nil {
			return clusterMapRet, apiServiceRet, nil
		}
	}
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
		req := new(api.CreateLeaseReq)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			htp.Error(w, http.StatusBadRequest, "error decoding JSON (%s)", err)
			return
		}

		clusterMap, svc, err := getSvcsAndClusters(clusterLister, services, gCloudProjID, gCloudZone, k8sServiceName)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error listing GKE clusters or talking to the k8s API (%s)", err)
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

		resp := api.CreateLeaseResp{
			KubeConfig:  kubeConfigStr,
			IP:          availableCluster.Endpoint,
			Token:       newToken.String(),
			ClusterName: availableCluster.Name,
		}

		leaseMap.CreateLease(newToken, leases.NewLease(availableCluster.Name, req.ExpirationTime(time.Now())))

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
