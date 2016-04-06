package k8s

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
