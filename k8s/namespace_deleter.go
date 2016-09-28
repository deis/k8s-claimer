package k8s

import (
	"k8s.io/client-go/1.4/pkg/api"
)

// NamespaceDeleter is a (k8s.io/kubernetes/pkg/client/unversioned).NamespaceInterface compatible
// interface designed only for deleting namespaces. It should be used as a parameter to functions
// so that they can be more easily unit tested
type NamespaceDeleter interface {
	Delete(name string, opts *api.DeleteOptions) error
}

// FakeNamespaceDeleter is a NamespaceDeleter implementation to be used in unit tests
type FakeNamespaceDeleter struct {
	NsDeleted []string
	Err       error
}

// Delete is the NamespaceDeleter interface implementation. It just returns f.Err
func (f *FakeNamespaceDeleter) Delete(name string, opts *api.DeleteOptions) error {
	f.NsDeleted = append(f.NsDeleted, name)
	return f.Err
}
