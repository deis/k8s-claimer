package gke

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"github.com/pborman/uuid"
)

var (
	errUnusedGKEClusterNotFound = errors.New("all GKE clusters are in use")
)

// Lease will lease an available cluster on GKE
func Lease(w http.ResponseWriter,
	req *api.CreateLeaseReq,
	clusterLister ClusterLister,
	services k8s.ServiceGetterUpdater,
	k8sServiceName,
	gCloudProjID,
	gCloudZone string) {

	clusterMap, svc, err := getSvcsAndClusters(clusterLister, services, gCloudProjID, gCloudZone, k8sServiceName)
	if err != nil {
		log.Printf("Error listing GKE clusters or talking to the k8s API -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error listing GKE clusters or talking to the k8s API -- %s", err)
	}

	leaseMap, err := leases.ParseMapFromAnnotations(svc.Annotations)
	if err != nil {
		log.Printf("Error parsing leases from Kubernetes annotations -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "error parsing leases from Kubernetes annotations -- %s", err)
	}

	availableCluster, err := searchForFreeCluster(clusterMap, leaseMap, req.ClusterRegex, req.ClusterVersion)
	if err != nil {
		switch e := err.(type) {
		case errNoAvailableOrExpiredClustersFound:
			log.Printf("No available clusters found")
			htp.Error(w, http.StatusConflict, "No available clusters found")
		case errExpiredLeaseGKEMissing:
			log.Printf("Cluster %s has an expired lease but doesn't exist in GKE", e.clusterName)
			htp.Error(w, http.StatusInternalServerError, "Cluster %s has an expired lease but doesn't exist in GKE", e.clusterName)
		default:
			log.Printf("Unknown error %s", e.Error())
			htp.Error(w, http.StatusInternalServerError, "Unknown error %s", e.Error())
		}
	}

	newToken := uuid.NewUUID()
	kubeConfig, err := k8s.CreateKubeConfigFromCluster(availableCluster)
	if err != nil {
		log.Printf("Error creating kubeconfig file for cluster %s -- %s", availableCluster.Name, err)
		htp.Error(w, http.StatusInternalServerError, "Error creating kubeconfig file for cluster %s -- %s", availableCluster.Name, err)
	}

	kubeConfigStr, err := k8s.MarshalAndEncodeKubeConfig(kubeConfig)
	if err != nil {
		log.Printf("Error marshaling & encoding kubeconfig -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error marshaling & encoding kubeconfig -- %s", err)
	}

	resp := api.CreateLeaseResp{
		KubeConfigStr:  kubeConfigStr,
		IP:             availableCluster.Endpoint,
		Token:          newToken.String(),
		ClusterName:    availableCluster.Name,
		ClusterVersion: availableCluster.CurrentNodeVersion,
	}

	leaseMap.CreateLease(newToken, leases.NewLease(availableCluster.Name, req.ExpirationTime(time.Now())))
	if err := k8s.SaveAnnotations(services, svc, leaseMap); err != nil {
		log.Printf("Error saving new lease to Kubernetes annotations -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error saving new lease to Kubernetes annotations -- %s", err)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding json -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error encoding json -- %s", err)
	}
}

func getSvcsAndClusters(
	clusterLister ClusterLister,
	services k8s.ServiceGetterUpdater,
	gCloudProjID,
	gCloudZone,
	k8sServiceName string,
) (*Map, *v1.Service, error) {

	errCh := make(chan error)
	doneCh := make(chan struct{})
	clusterMapCh := make(chan *Map)
	apiServiceCh := make(chan *v1.Service)
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
		clusterMap, err := ParseMapFromGKE(clusterLister, gCloudProjID, gCloudZone)
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

	var clusterMapRet *Map
	var apiServiceRet *v1.Service
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
