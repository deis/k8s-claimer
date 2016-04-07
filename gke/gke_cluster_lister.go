package gke

import (
	container "google.golang.org/api/container/v1"
)

// GKEClusterLister is a ClusterLister implementation that uses the GKE Go SDK to list clusters
// on a live GKE cluster
type GKEClusterLister struct {
	svc *container.Service
}

// NewGKEClusterLister creates a new GKEClusterLister configured to use the given client.
// See GetContainerService for how to create a new client.
func NewGKEClusterLister(svc *container.Service) *GKEClusterLister {
	return &GKEClusterLister{svc: svc}
}

// List is the ClusterLister interface implementation
func (g *GKEClusterLister) List(projectID, zone string) (*container.ListClustersResponse, error) {
	return g.svc.Projects.Zones.Clusters.List(projectID, zone).Do()
}
