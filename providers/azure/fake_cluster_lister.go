package azure

import "github.com/Azure/azure-sdk-for-go/arm/containerservice"

// FakeClusterLister is a ClusterLister implementation for use in unit tests
type FakeClusterLister struct {
	Resp *containerservice.ListResult
	Err  error
}

// List is the ClusterLister interface implementation. It just returns f.Resp, f.Err
func (f FakeClusterLister) List() (*containerservice.ListResult, error) {
	return f.Resp, f.Err
}

// NewFakeClusterLister returns a new FakeClusterLister
func NewFakeClusterLister(resp *containerservice.ListResult, err error) *FakeClusterLister {
	return &FakeClusterLister{
		Resp: resp,
		Err:  err,
	}
}
