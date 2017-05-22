package gke

import container "google.golang.org/api/container/v1"

// FakeClusterLister is a ClusterLister implementation for use in unit tests
type FakeClusterLister struct {
	Resp *container.ListClustersResponse
	Err  error
}

// List is the ClusterLister interface implementation. It just returns f.Resp, f.Err
func (f FakeClusterLister) List(projectID, zone string) (*container.ListClustersResponse, error) {
	return f.Resp, f.Err
}

// NewFakeClusterLister returns a new FakeClusterLister
func NewFakeClusterLister(resp *container.ListClustersResponse, err error) *FakeClusterLister {
	return &FakeClusterLister{
		Resp: resp,
		Err:  err,
	}
}
