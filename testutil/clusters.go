package testutil

import (
	container "google.golang.org/api/container/v1"
)

// GetClusters gets a list of clusters with names and versions defined
func GetClusters() []*container.Cluster {
	return []*container.Cluster{
		&container.Cluster{Name: "cluster1",
			CurrentNodeVersion: "9.9.9",
			Endpoint:           "192.168.1.1",
			MasterAuth:         &container.MasterAuth{}},
		&container.Cluster{Name: "cluster2",
			CurrentNodeVersion: "9.9.9",
			Endpoint:           "192.168.1.1",
			MasterAuth:         &container.MasterAuth{}},
		&container.Cluster{Name: "cluster3",
			CurrentNodeVersion: "9.9.9",
			Endpoint:           "192.168.1.1",
			MasterAuth:         &container.MasterAuth{}},
		&container.Cluster{Name: "cluster4",
			CurrentNodeVersion: "9.9.9",
			Endpoint:           "192.168.1.1",
			MasterAuth:         &container.MasterAuth{}},
		&container.Cluster{Name: "getClusterByVersion",
			CurrentNodeVersion: "1.1.1",
			Endpoint:           "192.168.1.1",
			MasterAuth:         &container.MasterAuth{}},
		&container.Cluster{Name: "getClusterByName",
			CurrentNodeVersion: "2.2.2",
			Endpoint:           "192.168.1.1",
			MasterAuth:         &container.MasterAuth{}},
	}
}
