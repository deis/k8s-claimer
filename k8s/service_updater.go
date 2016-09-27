package k8s

import (
	"k8s.io/client-go/1.4/pkg/api/v1"
)

// ServiceUpdater is a (k8s.io/kubernetes/pkg/client/unversioned).ServiceInterface compatible
// interface designed only for updating services. It should be used as a parameter to functions
// so that they can be more easily unit tested
type ServiceUpdater interface {
	Update(srv *v1.Service) (*v1.Service, error)
}

// FakeServiceUpdater is a ServiceUpdater implementation to be used in unit tests
type FakeServiceUpdater struct {
	RetSvc *v1.Service
	Err    error
}

// Update is the ServiceUpdater interface implementation. It just returns f.RetSvc, f.Err
func (f *FakeServiceUpdater) Update(srv *v1.Service) (*v1.Service, error) {
	return f.RetSvc, f.Err
}
