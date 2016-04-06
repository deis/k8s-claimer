package k8s

import (
	"k8s.io/kubernetes/pkg/api"
)

// ServiceUpdater is a (k8s.io/kubernetes/pkg/client/unversioned).ServiceInterface compatible
// interface designed only for updating services. It should be used as a parameter to functions
// so that they can be more easily unit tested
type ServiceUpdater interface {
	Update(srv *api.Service) (*api.Service, error)
}

// FakeServiceUpdater is a ServiceUpdater implementation to be used in unit tests
type FakeServiceUpdater struct {
	RetSvc *api.Service
	Err    error
}

// Update is the ServiceUpdater interface implementation. It just returns f.RetSvc, f.Err
func (f *FakeServiceUpdater) Update(srv *api.Service) (*api.Service, error) {
	return f.RetSvc, f.Err
}
