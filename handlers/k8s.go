package handlers

import (
	"errors"
	"time"

	"github.com/deis/k8s-claimer/leases"
	"k8s.io/kubernetes/pkg/api"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	errUnusedGKEClusterNotFound = errors.New("unused GKE cluster not found")
	errNoExpiredLeases          = errors.New("no expired leases exist")
)

// findExpiredLease searches in the leases in the svc annotations and returns the cluster name of
// the first expired lease it finds. If none found, returns an empty string and errNoExpiredLeases
func findExpiredLease(leaseMap *leases.Map) (*leases.UUIDAndLease, error) {
	now := time.Now()
	uuids, err := leaseMap.UUIDs()
	if err != nil {
		return nil, err
	}
	for _, u := range uuids {
		lease, _ := leaseMap.LeaseForUUID(u)
		exprTime, err := lease.ExpirationTime()
		if err != nil {
			return nil, err
		}
		if now.After(exprTime) {
			return leases.NewUUIDAndLease(u, lease), nil
		}
	}
	return nil, errNoExpiredLeases
}

func saveAnnotation(services k8s.ServiceInterface, svc *api.Service, leaseMap *leases.Map) error {
	return nil
}
