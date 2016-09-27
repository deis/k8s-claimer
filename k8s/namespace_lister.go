package k8s

import (
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/v1"
)

// NamespaceLister is a (k8s.io/kubernetes/pkg/client/unversioned).NamespaceInterface compatible
// interface designed only for listing namespaces. It should be used as a parameter to functions
// so that they can be more easily unit tested
type NamespaceLister interface {
	List(opts api.ListOptions) (*v1.NamespaceList, error)
}

// FakeNamespaceLister is a NamespaceLister implementation to be used in unit tests
type FakeNamespaceLister struct {
	NsList *v1.NamespaceList
	Err    error
}

// List is the NamespaceLister interface implementation. It just returns f.NsList, f.Err
func (f FakeNamespaceLister) List(opts api.ListOptions) (*v1.NamespaceList, error) {
	return f.NsList, f.Err
}
