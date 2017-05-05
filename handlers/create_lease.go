package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/providers/azure"
	"github.com/deis/k8s-claimer/providers/gke"
)

// CreateLease creates the handler that responds to the POST /lease endpoint
func CreateLease(
	services k8s.ServiceGetterUpdater,
	k8sServiceName string,
	gkeClusterLister gke.ClusterLister,
	azureClusterLister azure.ClusterLister,
	azureConfig *config.Azure,
	googleConfig *config.Google,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := new(api.CreateLeaseReq)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			log.Printf("Error decoding JSON -- %s", err)
			htp.Error(w, http.StatusBadRequest, "Error decoding JSON -- %s", err)
			return
		}

		switch req.CloudProvider {
		case "google":
			gke.Lease(w, req, gkeClusterLister, services, k8sServiceName, googleConfig.ProjectID, googleConfig.Zone)
		case "azure":
			azure.Lease(w, req, azureClusterLister, services, azureConfig, k8sServiceName)
		default:
			log.Printf("Unable to find suitable provider for this request -- Provider:%s", req.CloudProvider)
			htp.Error(w, http.StatusBadRequest, "Unable to find suitable provider for this request -- Provider:%s", req.CloudProvider)
		}
	})
}
