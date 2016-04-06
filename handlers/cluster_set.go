package handlers

import (
	container "google.golang.org/api/container/v1"
)

type clusterSet map[string]*container.Cluster

func clusterSetFromGKE(containerSvc *container.Service, projID, zone string) (clusterSet, error) {
	ret := make(map[string]*container.Cluster)
	clustersResp, err := containerSvc.Projects.Zones.Clusters.List(projID, zone).Do()
	if err != nil {
		return nil, err
	}
	for _, cluster := range clustersResp.Clusters {
		ret[cluster.Name] = cluster
	}
	return ret, nil
}

func (c clusterSet) getByName(name string) (*container.Cluster, bool) {
	cl, found := c[name]
	return cl, found
}
