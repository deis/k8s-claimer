package handlers

import (
	"net/http"

	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"github.com/pborman/uuid"
)

// DeleteLease returns the http handler for the DELETE /lease/{token} endpoint
func DeleteLease(services k8s.ServiceGetterUpdater, k8sServiceName string, namespaces k8s.NamespaceListerDeleter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathElts := htp.SplitPath(r)
		if len(pathElts) != 2 {
			htp.Error(w, http.StatusBadRequest, "path must be in the format /lease/{token}")
			return
		}
		leaseToken := uuid.Parse(pathElts[1])
		if leaseToken == nil {
			htp.Error(w, http.StatusBadRequest, "lease token %s is invalid", pathElts[1])
			return
		}

		svc, err := services.Get(k8sServiceName)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error getting the %s service (%s)", k8sServiceName, err)
			return
		}
		leaseMap, err := leases.ParseMapFromAnnotations(svc.Annotations)
		if err != nil {
			htp.Error(w, http.StatusInternalServerError, "error getting annotations for the %s service (%s)", k8sServiceName, err)
			return
		}
		// blow away the lease, regardless of whether it's expired or not. the create endpoint deletes
		// the lease from annotations, replacing the lease for a cluster with a new UUID anyway
		deleted := leaseMap.DeleteLease(leaseToken)
		if !deleted {
			htp.Error(w, http.StatusConflict, "lease %s doesn't exist", leaseToken)
			return
		}
		if err := saveAnnotations(services, svc, leaseMap); err != nil {
			htp.Error(w, http.StatusInternalServerError, "error saving new annotations (%s)", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
