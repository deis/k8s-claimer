package k8s

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
