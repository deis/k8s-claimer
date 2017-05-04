package k8s

import (
	"k8s.io/client-go/pkg/api/v1"
)

// NamespaceDeleter is a (k8s.io/kubernetes/pkg/client/unversioned).NamespaceInterface compatible
// interface designed only for deleting namespaces. It should be used as a parameter to functions
// so that they can be more easily unit tested
type NamespaceDeleter interface {
	Delete(name string, opts *v1.DeleteOptions) error
}

// FakeNamespaceDeleter is a NamespaceDeleter implementation to be used in unit tests
type FakeNamespaceDeleter struct {
	NsDeleted []string
	Err       error
}

// Delete is the NamespaceDeleter interface implementation. It just returns f.Err
func (f *FakeNamespaceDeleter) Delete(name string, opts *v1.DeleteOptions) error {
	f.NsDeleted = append(f.NsDeleted, name)
	return f.Err
}

// NewFakeNamespaceDeleter returns a new FakeNamespaceDeleter
func NewFakeNamespaceDeleter(err error) *FakeNamespaceDeleter {
	return &FakeNamespaceDeleter{Err: err}
}
