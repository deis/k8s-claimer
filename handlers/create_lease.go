package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/providers/gke"
)

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
			log.Printf("Error decoding JSON -- %s", err)
			htp.Error(w, http.StatusBadRequest, "Error decoding JSON -- %s", err)
			return
		}

		switch req.CloudProvider {
		case "google":
			gke.Lease(w, req, clusterLister, services, k8sServiceName, gCloudProjID, gCloudZone)
			return
		case "azure":
			// fork to azure leasing here
			return
		default:
			log.Printf("Unable to find suitable provider for this request -- Provider:%s", req.CloudProvider)
			htp.Error(w, http.StatusBadRequest, "Unable to find suitable provider for this request -- Provider:%s", req.CloudProvider)
		}
	})
}
