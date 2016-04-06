package handlers

import (
	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/leases"
	container "google.golang.org/api/container/v1"
)

type errNoClusterWithName struct {
	name string
}

func (e errNoClusterWithName) Error() string {
	return "no cluster with name " + e.name
}

// findUnusedGKECluster finds a GKE cluster that's not currently in use according to the
// annotations in svc. returns errUnusedGKEClusterNotFound if none is found
func findUnusedGKECluster(clusterMap *clusters.Map, leaseMap *leases.Map) (*container.Cluster, error) {
	clusterNames := clusterMap.Names()
	for _, clusterName := range clusterNames {
		cluster, _ := clusterMap.ClusterByName(clusterName)
		_, found := leaseMap.LeaseByClusterName(clusterName)
		if !found {
			return cluster, nil
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

func createKubeConfigFromCluster(cluster *container.Cluster) string {
	// TODO: implement
	return ""
}
