package clusters

import (
	"github.com/deis/k8s-claimer/gke"
	container "google.golang.org/api/container/v1"
)

// Map is a map from cluster name to GKE cluster
type Map struct {
	nameMap map[string]*container.Cluster
}

func clustersToMap(c []*container.Cluster) map[string]*container.Cluster {
	ret := make(map[string]*container.Cluster)
	for _, cluster := range c {
		ret[cluster.Name] = cluster
	}
	return ret
}

// ParseMapFromGKE calls the GKE API to get a list of clusters, then returns a Map representation
// of those clusters. Returns nil and an appropriate error if any errors occurred along the way
func ParseMapFromGKE(clusterLister gke.ClusterLister, projID, zone string) (*Map, error) {
	clustersResp, err := clusterLister.List(projID, zone)
	if err != nil {
		return nil, err
	}
	return &Map{nameMap: clustersToMap(clustersResp.Clusters)}, nil
}

// ClusterByName returns the cluster of the given cluster name. Returns nil and false if no
// cluster with the given name exists, non-nil and true otherwise
func (m Map) ClusterByName(name string) (*container.Cluster, bool) {
	cl, found := m.nameMap[name]
	return cl, found
}

// Names returns all cluster names in the map. The order of the returned slice is undefined
func (m Map) Names() []string {
	ret := make([]string, len(m.nameMap))
	i := 0
	for name := range m.nameMap {
		ret[i] = name
		i++
	}
	return ret
}
