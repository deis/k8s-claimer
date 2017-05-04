package k8s

import "k8s.io/client-go/pkg/api/v1"

// ServiceGetterUpdater contains both the listing and updating capabilities of a
// (k8s.io/kubernetes/pkg/client/unversioned).ServiceInterface
type ServiceGetterUpdater interface {
	ServiceGetter
	ServiceUpdater
}

// FakeServiceGetterUpdater is a ServiceGetterUpdater composed of both a FakeServiceGetter
// and a FakeServiceUpdater
type FakeServiceGetterUpdater struct {
	*FakeServiceGetter
	*FakeServiceUpdater
}

func NewFakeServiceGetterUpdater(
	getSvc *v1.Service,
	getErr error,
	updateSvc *v1.Service,
	updateErr error,
) *FakeServiceGetterUpdater {
	return &FakeServiceGetterUpdater{
		FakeServiceGetter:  NewFakeServiceGetter(getSvc, getErr),
		FakeServiceUpdater: NewFakeServiceUpdater(updateSvc, updateErr),
	}
}

func NewFakeServiceGetter(svc *v1.Service, err error) *FakeServiceGetter {
	return &FakeServiceGetter{Svc: svc, Err: err}
}

func NewFakeServiceUpdater(retSvc *v1.Service, err error) *FakeServiceUpdater {
	return &FakeServiceUpdater{RetSvc: retSvc, Err: err}
}
