package azure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/leases"
	"github.com/pborman/uuid"

	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/k8s"
)

// Lease will search for an available cluster on Azure which matches the parameters passed in on the request
// It will write back on the response the necessary connection information in json format
func Lease(w http.ResponseWriter,
	req *api.CreateLeaseReq,
	clusterLister ClusterLister,
	services k8s.ServiceGetterUpdater,
	azureConfig *config.Azure,
	k8sServiceName string) {

	clusterMap, svc, err := getSvcsAndClusters(clusterLister, services, k8sServiceName)
	if err != nil {
		log.Printf("Error listing Azure clusters or talking to the k8s API -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error listing Azure clusters or talking to the k8s API -- %s", err)
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
		case errExpiredLeaseAzureMissing:
			log.Printf("Cluster %s has an expired lease but doesn't exist in Azure", e.clusterName)
			htp.Error(w, http.StatusInternalServerError, "Cluster %s has an expired lease but doesn't exist in Azure", e.clusterName)
			return
		default:
			log.Printf("Unknown error %s", e.Error())
			htp.Error(w, http.StatusInternalServerError, "Unknown error %s", e.Error())
			return
		}
	}

	// There is currently no way to fetch the kubeconfig from the Azure API
	// So we must scp the file off the master node
	kubeConfig, err := FetchKubeConfig(*availableCluster.MasterProfile.Fqdn)
	if err != nil {
		log.Printf("Error creating kubeconfig file for cluster %s -- %s", *availableCluster.Name, err)
		htp.Error(w, http.StatusInternalServerError, "Error creating kubeconfig file for cluster %s -- %s", *availableCluster.Name, err)
		return
	}
	newToken := uuid.NewUUID()
	kubeConfigStr, err := k8s.MarshalAndEncodeKubeConfig(kubeConfig)
	if err != nil {
		log.Printf("Error marshaling & encoding kubeconfig -- %s", err)
		htp.Error(w, http.StatusInternalServerError, "Error marshaling & encoding kubeconfig -- %s", err)
		return
	}

	resp := api.CreateLeaseResp{
		KubeConfigStr: kubeConfigStr,
		IP:            *availableCluster.MasterProfile.Fqdn,
		Token:         newToken.String(),
		ClusterName:   *availableCluster.Name,
	}

	leaseMap.CreateLease(newToken, leases.NewLease(*availableCluster.Name, req.ExpirationTime(time.Now())))
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

func getSvcsAndClusters(clusterLister ClusterLister, services k8s.ServiceGetterUpdater, k8sServiceName string) (*Map, *v1.Service, error) {

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
		clusterMap, err := ParseMapFromAzure(clusterLister)
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

// FetchKubeConfig will scp the kubeconfig file from the master server into /tmp/kubeconfig<tempfile>
func FetchKubeConfig(server string) (*k8s.KubeConfig, error) {
	f, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		log.Printf("Error while trying to create temp file for kubeconfig:%s\n", err)
		return nil, err
	}
	conn := fmt.Sprintf("azureuser@%s:.kube/config", server)
	cmd := exec.Command("scp", "-oStrictHostKeyChecking=no", conn, f.Name())
	err := cmd.Start()
	if err != nil {
		log.Printf("Error while trying to start SCP:%s\n", err)
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Error while waiting for SCP to complete:%s\n", err)
		return nil, err
	}

	contentInBytes, err := ioutil.ReadFile(f.Name())
	kubeConfig, err := k8s.CreateKubeConfig(contentInBytes)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}
