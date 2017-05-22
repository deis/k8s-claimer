package gke

import (
	container "google.golang.org/api/container/v1"
)

// ClusterLister is an interface for listing GKE clusters. It has an adapter for the
// standard *(google.golang.org/api/container/v1).Service as well as a fake implementation,
// to be used in unit tests. Use this as a parameter in your funcs so that they can be more
// easily unit tested
type ClusterLister interface {
	// List lists all of the clusters in the given project and zone
	List(projectID, zone string) (*container.ListClustersResponse, error)
}
