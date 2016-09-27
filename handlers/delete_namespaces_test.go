package handlers

import (
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/k8s"
	"k8s.io/client-go/1.4/pkg/api/v1"
)

func getNSListerDeleter(listedNamespaces []string) *k8s.FakeNamespaceListerDeleter {
	ret := &k8s.FakeNamespaceListerDeleter{
		FakeNamespaceLister: &k8s.FakeNamespaceLister{
			NsList: &v1.NamespaceList{},
		},
		FakeNamespaceDeleter: &k8s.FakeNamespaceDeleter{},
	}
	for _, listedNamespace := range listedNamespaces {
		ret.FakeNamespaceLister.NsList.Items = append(ret.FakeNamespaceLister.NsList.Items, v1.Namespace{
			ObjectMeta: v1.ObjectMeta{Name: listedNamespace},
		})
	}
	return ret
}

func TestDeleteNamespaces(t *testing.T) {
	defaultSkip := []string{"kube-system", "default"}
	nsList := []string{"ns1", "ns2", "default", "kube-system"}

	// test deleting all namespaces
	nsListerDeleter := getNSListerDeleter(nsList)
	skip := make(map[string]struct{})
	assert.NoErr(t, deleteNamespaces(nsListerDeleter, skip))
	assert.Equal(t, len(nsListerDeleter.FakeNamespaceDeleter.NsDeleted), len(nsList)-len(defaultSkip), "number of deleted items")

	// test skipping deletion of some namespaces
	nsListerDeleter = getNSListerDeleter(nsList)
	skip = map[string]struct{}{"ns1": struct{}{}}
	assert.NoErr(t, deleteNamespaces(nsListerDeleter, skip))
	remaining := map[string]struct{}{}
	for _, deletedNS := range nsListerDeleter.FakeNamespaceDeleter.NsDeleted {
		remaining[deletedNS] = struct{}{}
	}
	numDeleted := len(nsListerDeleter.FakeNamespaceDeleter.NsDeleted)
	numSkipped := len(defaultSkip) + len(skip)
	assert.Equal(t, numDeleted, len(nsList)-numSkipped, "number of remaining namespaces")
}
