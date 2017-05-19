package gke

import (
	"context"
	"encoding/json"
	"fmt"
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

// Lease will search for an available cluster on GKE which matches the parameters passed in on the request
// It will write back on the response the necessary connection information in json format
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
		return
	}

	leaseMap, err := leases.ParseMapFromAnnotations(svc.Annotations)
	if err != nil {
		log.Printf("Error parsing leases from Kubernetes annotations -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "error parsing leases from Kubernetes annotations -- %s", err)
		return
	}

	availableCluster, err := searchForFreeCluster(clusterMap, leaseMap, req.ClusterRegex, req.ClusterVersion)
	if err != nil {
		switch e := err.(type) {
		case errNoAvailableOrExpiredClustersFound:
			log.Printf("No available clusters found")
			htp.Error(w, http.StatusConflict, "No available clusters found")
			return
		case errExpiredLeaseGKEMissing:
			log.Printf("Cluster %s has an expired lease but doesn't exist in GKE", e.clusterName)
			htp.Error(w, http.StatusInternalServerError, "Cluster %s has an expired lease but doesn't exist in GKE", e.clusterName)
			return
		default:
			log.Printf("Unknown error %s", e.Error())
			htp.Error(w, http.StatusInternalServerError, "Unknown error %s", e.Error())
			return
		}
	}

	newToken := uuid.NewUUID()
	kubeConfig, err := k8s.CreateKubeConfigFromCluster(availableCluster)
	if err != nil {
		log.Printf("Error creating kubeconfig file for cluster %s -- %s", availableCluster.Name, err)
		htp.Error(w, http.StatusInternalServerError, "Error creating kubeconfig file for cluster %s -- %s", availableCluster.Name, err)
		return
	}

	kubeConfigStr, err := k8s.MarshalAndEncodeKubeConfig(kubeConfig)
	if err != nil {
		log.Printf("Error marshaling & encoding kubeconfig -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error marshaling & encoding kubeconfig -- %s", err)
		return
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
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding json -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error encoding json -- %s", err)
		return
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
	clusterMapCh := make(chan *Map)
	apiServiceCh := make(chan *v1.Service)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func() {
		svc, err := services.Get(k8sServiceName)
		if err != nil {
			select {
			case errCh <- err:
			}
			return
		}
		select {
		case apiServiceCh <- svc:
		case <-ctx.Done():
			errCh <- fmt.Errorf("Timeout exceeded while trying to fetch services from Kubernetes API")
		}
	}()
	go func() {
		clusterMap, err := ParseMapFromGKE(clusterLister, gCloudProjID, gCloudZone)
		if err != nil {
			select {
			case errCh <- err:
			case <-ctx.Done():
				errCh <- fmt.Errorf("Timeout exceeded while trying to fetch clusters from Google API")
			}
			return
		}
		select {
		case clusterMapCh <- clusterMap:
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
