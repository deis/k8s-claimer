package handlers

import (
	"fmt"

	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/leases"
	container "google.golang.org/api/container/v1"
)

type errNoAvailableOrExpiredClustersFound struct{}

func (e errNoAvailableOrExpiredClustersFound) Error() string {
	return "no available or expired clusters found"
}

type errExpiredLeaseGKEMissing struct {
	clusterName string
}

func (e errExpiredLeaseGKEMissing) Error() string {
	return fmt.Sprintf("cluster %s has an expired lease but does not exist in GKE", e.clusterName)
}

// searchForFreeCluster looks for an available GKE cluster to lease.
// It will try and match clusterRegex if possible
//
// Returns errNoAvailableOrExpiredClustersFound if it found no free or expired lease clusters.
// Returns errExpiredLeaseGKEMissing if it found an expired lease but the cluster associated with
// that lease doesn't exist in GKE
func searchForFreeCluster(clusterMap *clusters.Map, leaseMap *leases.Map, clusterRegex string) (*container.Cluster, error) {
	unusedCluster, unusedClusterErr := findUnusedGKECluster(clusterMap, leaseMap, clusterRegex)
	uuidAndLease, expiredLeaseErr := findExpiredLease(leaseMap)
	if unusedClusterErr != nil && expiredLeaseErr != nil {
		return nil, errNoAvailableOrExpiredClustersFound{}
	}
	var availableCluster *container.Cluster
	if unusedCluster != nil {
		availableCluster = unusedCluster
	}
	if expiredLeaseErr == nil {
		cl, ok := clusterMap.ClusterByName(uuidAndLease.Lease.ClusterName)
		if !ok {
			return nil, errExpiredLeaseGKEMissing{clusterName: uuidAndLease.Lease.ClusterName}
		}
		leaseMap.DeleteLease(uuidAndLease.UUID)
		availableCluster = cl
	}

	return availableCluster, nil
}
