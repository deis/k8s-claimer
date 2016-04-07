package k8s

import (
	"k8s.io/kubernetes/pkg/api"
)

// ServiceGetter is a (k8s.io/kubernetes/pkg/client/unversioned).ServiceInterface compatible
// interface designed only for getting a service. It should be used as a parameter to functions
// so that they can be more easily unit tested
type ServiceGetter interface {
	Get(name string) (*api.Service, error)
}

// FakeServiceGetter is a ServiceGetter implementation to be used in unit tests
type FakeServiceGetter struct {
	Svc *api.Service
	Err error
}

// Get is the ServiceGetter interface implementation. It just returns f.Svc, f.Err
func (f FakeServiceGetter) Get(name string) (*api.Service, error) {
	return f.Svc, f.Err
}
