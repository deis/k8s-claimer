package handlers

import (
	"fmt"

	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/leases"
	container "google.golang.org/api/container/v1"
)

type errNoSuchCluster struct {
	name string
}

func (e errNoSuchCluster) Error() string {
	return fmt.Sprintf("no such cluster %s", e.name)
}

func getClusterFromLease(lease *leases.Lease, clusterLister gke.ClusterLister, projID, zone string) (*container.Cluster, error) {
	clusterMap, err := clusters.ParseMapFromGKE(clusterLister, projID, zone)
	if err != nil {
		return nil, err
	}
	cl, exists := clusterMap.ClusterByName(lease.ClusterName)
	if !exists {
		return nil, errNoSuchCluster{name: lease.ClusterName}
	}

	return cl, nil
}
