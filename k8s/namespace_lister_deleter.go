package k8s

import "k8s.io/client-go/pkg/api/v1"

// NamespaceListerDeleter contains both the listing and deleting capabilities of a
// (k8s.io/kubernetes/pkg/client/unversioned).NamespaceInterface
type NamespaceListerDeleter interface {
	NamespaceLister
	NamespaceDeleter
}

// FakeNamespaceListerDeleter is a NamespaceListerDeleter composed of both a FakeNamespaceLister
// and a FakeNamespaceDeleter
type FakeNamespaceListerDeleter struct {
	*FakeNamespaceLister
	*FakeNamespaceDeleter
}

// GetNSFunc returns a function for NamespaceListerDeleter
func GetNSFunc(nsListerDeleter NamespaceListerDeleter, err error) func(*KubeConfig) (NamespaceListerDeleter, error) {
	return func(*KubeConfig) (NamespaceListerDeleter, error) {
		return nsListerDeleter, err
	}
}

// NewFakeNamespaceListerDeleter returns a new FakeNamespaceListerDeleter
func NewFakeNamespaceListerDeleter(listNs *v1.NamespaceList, listErr error, deleteErr error) *FakeNamespaceListerDeleter {
	return &FakeNamespaceListerDeleter{
		FakeNamespaceLister:  NewFakeNamespaceLister(listNs, listErr),
		FakeNamespaceDeleter: NewFakeNamespaceDeleter(deleteErr),
	}
}
