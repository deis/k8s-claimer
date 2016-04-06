package handlers

import (
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
func findUnusedGKECluster(clusterSet clusterSet, annotations map[string]string) (*container.Cluster, error) {
	existingLeases := getLeasesFromAnnotations(annotations)
	for clusterName, cluster := range clusterSet {
		_, found := existingLeases[clusterName]
		if !found {
			return cluster, nil
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

func createKubeConfigFromCluster(cluster *container.Cluster) string {
	return ""
}

// findClusterByName finds the cluster of the given name in clusterName. returns nil and an
// errNoClusterWithName error if no cluster with the given name was found
func findClusterByName(clusterName string, clusterSet clusterSet) (*container.Cluster, error) {
	cluster, found := clusterSet.getByName(clusterName)
	if !found {
		return nil, errNoClusterWithName{name: clusterName}
	}
	return cluster, nil
}
