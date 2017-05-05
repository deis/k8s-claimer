package testutil

import (
	"github.com/Azure/azure-sdk-for-go/arm/containerservice"
	container "google.golang.org/api/container/v1"
)

// GetGKEClusters gets a list of clusters with names and versions defined
func GetGKEClusters() []*container.Cluster {
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

// GetAzureClusters gets a list of clusters with names and versions defined
func GetAzureClusters() *[]containerservice.ContainerService {
	cluster1 := "cluster1"
	cluster2 := "cluster2"
	cluster3 := "cluster3"
	cluster4 := "cluster4"
	getClusterByName := "getClusterByName"
	getClusterByVersion := "getClusterByVersion"
	return &[]containerservice.ContainerService{
		containerservice.ContainerService{
			ID:   &cluster1,
			Name: &cluster1,
		},
		containerservice.ContainerService{
			ID:   &cluster2,
			Name: &cluster2,
		},
		containerservice.ContainerService{
			ID:   &cluster3,
			Name: &cluster3,
		},
		containerservice.ContainerService{
			ID:   &cluster4,
			Name: &cluster4,
		},
		containerservice.ContainerService{
			ID:   &getClusterByName,
			Name: &getClusterByName,
		},
		containerservice.ContainerService{
			ID:   &getClusterByVersion,
			Name: &getClusterByVersion,
		},
	}
}
