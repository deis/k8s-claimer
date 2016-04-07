package k8s

import (
	"k8s.io/kubernetes/pkg/api"
)

// ServiceLister is a (k8s.io/kubernetes/pkg/client/unversioned).ServiceInterface compatible
// interface designed only for listing services. It should be used as a parameter to functions
// so that they can be more easily unit tested
type ServiceLister interface {
	List(opts api.ListOptions) (*api.ServiceList, error)
}

// FakeServiceLister is a ServiceLister implementation to be used in unit tests
type FakeServiceLister struct {
	SvcList *api.ServiceList
	Err     error
}

// List is the ServiceLister interface implementation. It just returns f.SvcList, f.Err
func (f FakeServiceLister) List(opts api.ListOptions) (*api.ServiceList, error) {
	return f.SvcList, f.Err
}
