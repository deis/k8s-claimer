package k8s

import (
	"k8s.io/kubernetes/pkg/api"
)

// NamespaceLister is a (k8s.io/kubernetes/pkg/client/unversioned).NamespaceInterface compatible
// interface designed only for listing namespaces. It should be used as a parameter to functions
// so that they can be more easily unit tested
type NamespaceLister interface {
	List(opts api.ListOptions) (*api.NamespaceList, error)
}

// FakeNamespaceLister is a NamespaceLister implementation to be used in unit tests
type FakeNamespaceLister struct {
	NsList *api.NamespaceList
	Err    error
}

// List is the NamespaceLister interface implementation. It just returns f.NsList, f.Err
func (f FakeNamespaceLister) List(opts api.ListOptions) (*api.NamespaceList, error) {
	return f.NsList, f.Err
}
