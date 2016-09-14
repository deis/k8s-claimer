package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
)

func getNSFunc(nsListerDeleter k8s.NamespaceListerDeleter, err error) func(*k8s.KubeConfig) (k8s.NamespaceListerDeleter, error) {
	return func(*k8s.KubeConfig) (k8s.NamespaceListerDeleter, error) {
		return nsListerDeleter, err
	}
}

func TestDeleteLeaseNoToken(t *testing.T) {
	getterUpdater := newFakeServiceGetterUpdater(nil, nil, nil, nil)
	clusterLister := newFakeClusterLister(newListClusterResp(testutil.GetClusters()), nil)
	nsListerDeleter := newFakeNamespaceListerDeleter(&api.NamespaceList{}, nil, nil)
	hdl := DeleteLease(getterUpdater, clusterLister, "claimer", "proj1", "zone1", true, getNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease", nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidLeaseToken(t *testing.T) {
	getterUpdater := newFakeServiceGetterUpdater(nil, nil, nil, nil)
	clusterLister := newFakeClusterLister(newListClusterResp(testutil.GetClusters()), nil)
	nsListerDeleter := newFakeNamespaceListerDeleter(&api.NamespaceList{}, nil, nil)
	hdl := DeleteLease(getterUpdater, clusterLister, "claimer", "proj1", "zone1", true, getNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease/abcd", nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidAnnotations(t *testing.T) {
	// Issue a DELETE with a lease token that doesn't point to a valid lease.
	// Annotations have invalid data in them, which the lease parser should just ignore.
	getterUpdater := newFakeServiceGetterUpdater(
		&api.Service{ObjectMeta: api.ObjectMeta{Annotations: map[string]string{"a": "b"}}},
		nil,
		nil,
		nil,
	)
	clusterLister := newFakeClusterLister(newListClusterResp(testutil.GetClusters()), nil)
	nsListerDeleter := newFakeNamespaceListerDeleter(&api.NamespaceList{}, nil, nil)
	hdl := DeleteLease(getterUpdater, clusterLister, "claimer", "proj1", "zone1", true, getNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease/"+uuid.New(), nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusConflict, "response code")
}

func TestDeleteLeaseNoSuchLease(t *testing.T) {
	getterUpdater := newFakeServiceGetterUpdater(
		&api.Service{ObjectMeta: api.ObjectMeta{Annotations: map[string]string{}}},
		nil,
		nil,
		nil,
	)
	clusterLister := newFakeClusterLister(newListClusterResp(testutil.GetClusters()), nil)
	nsListerDeleter := newFakeNamespaceListerDeleter(&api.NamespaceList{}, nil, nil)
	hdl := DeleteLease(getterUpdater, clusterLister, "claimer", "proj1", "zone1", true, getNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease/"+uuid.New(), nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusConflict, "response code")
}

func TestDeleteLeaseExists(t *testing.T) {
	leaseableClusters := testutil.GetClusters()
	uuids := make([]uuid.UUID, len(leaseableClusters))
	annos := testutil.GetRawAnnotations(
		leaseableClusters,
		leases.TimeFormat,
		func(i int) time.Time {
			return time.Now().Add(1 * time.Hour)
		},
		func(i int) uuid.UUID {
			ret := uuid.NewUUID()
			uuids[i] = ret
			return ret
		},
	)
	for i, u := range uuids {
		getterUpdater := newFakeServiceGetterUpdater(
			&api.Service{ObjectMeta: api.ObjectMeta{Annotations: annos}},
			nil,
			nil,
			nil,
		)
		path := "/lease/" + u.String()

		defaultNamespaces := []string{"default", "kube-system"}
		namespaces := append(defaultNamespaces, "ns1", "ns2")
		nsList := api.NamespaceList{}
		for _, namespace := range namespaces {
			nsList.Items = append(nsList.Items, api.Namespace{ObjectMeta: api.ObjectMeta{Name: namespace}})
		}

		listClusterResp := newListClusterResp(testutil.GetClusters())
		clusterLister := newFakeClusterLister(listClusterResp, nil)
		nsListerDeleter := newFakeNamespaceListerDeleter(&nsList, nil, nil)
		hdl := DeleteLease(getterUpdater, clusterLister, "claimer", "proj1", "zone1", true, getNSFunc(nsListerDeleter, nil))
		req, err := http.NewRequest("DELETE", path, nil)
		req.Header.Set("Authorization", "some awesome token")
		if err != nil {
			t.Errorf("trial %d: error creating request %s (%s)", i, path, err)
			continue
		}
		res := httptest.NewRecorder()
		hdl.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("trial %d: status code for path %s was %d, not %d", i, path, res.Code, http.StatusOK)
			continue
		}
		// ensure namespaces were deleted
		assert.Equal(t, len(nsListerDeleter.NsDeleted), len(namespaces)-len(skipDeleteNamespaces), "number of deleted namespaces")
		nsMap := map[string]struct{}{}
		for _, ns := range namespaces {
			nsMap[ns] = struct{}{}
		}
		for _, deletedNS := range nsListerDeleter.NsDeleted {
			if _, ok := nsMap[deletedNS]; !ok {
				t.Errorf("namespace %s was deleted but wasn't in the original namespace list", deletedNS)
			}
		}
	}
}

func TestDeleteLeaseExistsNoClearNamespaces(t *testing.T) {
	leaseableClusters := testutil.GetClusters()
	uuids := make([]uuid.UUID, len(leaseableClusters))
	annos := testutil.GetRawAnnotations(
		leaseableClusters,
		leases.TimeFormat,
		func(i int) time.Time {
			return time.Now().Add(1 * time.Hour)
		},
		func(i int) uuid.UUID {
			ret := uuid.NewUUID()
			uuids[i] = ret
			return ret
		},
	)
	for i, u := range uuids {
		getterUpdater := newFakeServiceGetterUpdater(
			&api.Service{ObjectMeta: api.ObjectMeta{Annotations: annos}},
			nil,
			nil,
			nil,
		)
		path := "/lease/" + u.String()

		defaultNamespaces := []string{"default", "kube-system"}
		namespaces := append(defaultNamespaces, "ns1", "ns2")
		nsList := api.NamespaceList{}
		for _, namespace := range namespaces {
			nsList.Items = append(nsList.Items, api.Namespace{ObjectMeta: api.ObjectMeta{Name: namespace}})
		}

		listClusterResp := newListClusterResp(leaseableClusters)
		clusterLister := newFakeClusterLister(listClusterResp, nil)
		nsListerDeleter := newFakeNamespaceListerDeleter(&nsList, nil, nil)
		hdl := DeleteLease(getterUpdater, clusterLister, "claimer", "proj1", "zone1", false, getNSFunc(nsListerDeleter, nil))
		req, err := http.NewRequest("DELETE", path, nil)
		req.Header.Set("Authorization", "some awesome token")
		if err != nil {
			t.Errorf("trial %d: error creating request %s (%s)", i, path, err)
			continue
		}
		res := httptest.NewRecorder()
		hdl.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("trial %d: status code for path %s was %d, not %d", i, path, res.Code, http.StatusOK)
			continue
		}
		// ensure namespaces were not deleted
		assert.Equal(t, len(nsListerDeleter.NsDeleted), 0, "number of deleted namespaces")
	}
}
