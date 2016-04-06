package handlers

import (
	container "google.golang.org/api/container/v1"
)

func getGKEClusters(containerSvc *container.Service, projID, zone string) ([]*container.Cluster, error) {
	clustersResp, err := containerSvc.Projects.Zones.Clusters.List().Do()
	if err != nil {
		return nil, err
	}
	return clustersResp.Clusters, nil
}
