package clusters

import (
	container "google.golang.org/api/container/v1"
)

// Map is a map from cluster name to GKE cluster
type Map struct {
	nameMap map[string]*container.Cluster
}

// ParseMapFromGKE calls the GKE API to get a list of clusters, then returns a map representation
// of those clusters. Returns nil and an appropriate error if any errors occurred along the way
func ParseMapFromGKE(containerSvc *container.Service, projID, zone string) (*Map, error) {
	ret := make(map[string]*container.Cluster)
	clustersResp, err := containerSvc.Projects.Zones.Clusters.List(projID, zone).Do()
	if err != nil {
		return nil, err
	}
	for _, cluster := range clustersResp.Clusters {
		ret[cluster.Name] = cluster
	}
	return &Map{nameMap: ret}, nil
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
