package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/providers/gke"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
)

func TestDeleteLeaseNoToken(t *testing.T) {
	getterUpdater := k8s.NewFakeServiceGetterUpdater(nil, nil, nil, nil)
	clusterLister := gke.NewFakeClusterLister(newListClusterResp(testutil.GetGKEClusters()), nil)
	nsListerDeleter := k8s.NewFakeNamespaceListerDeleter(&v1.NamespaceList{}, nil, nil)
	googleConfig := &config.Google{AccountFileJSON: "test", ProjectID: "proj1", Zone: "zone1"}
	hdl := DeleteLease(getterUpdater, "claimer", clusterLister, nil, nil, googleConfig, true, k8s.GetNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease", nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidLeaseToken(t *testing.T) {
	getterUpdater := k8s.NewFakeServiceGetterUpdater(nil, nil, nil, nil)
	clusterLister := gke.NewFakeClusterLister(newListClusterResp(testutil.GetGKEClusters()), nil)
	nsListerDeleter := k8s.NewFakeNamespaceListerDeleter(&v1.NamespaceList{}, nil, nil)
	googleConfig := config.Google{ProjectID: "proj1", Zone: "zone1"}
	hdl := DeleteLease(getterUpdater, "claimer", clusterLister, nil, nil, &googleConfig, true, k8s.GetNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease/google/abcd", nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidAnnotations(t *testing.T) {
	// Issue a DELETE with a lease token that doesn't point to a valid lease.
	// Annotations have invalid data in them, which the lease parser should just ignore.
	getterUpdater := k8s.NewFakeServiceGetterUpdater(
		&v1.Service{ObjectMeta: v1.ObjectMeta{Annotations: map[string]string{"a": "b"}}},
		nil,
		nil,
		nil,
	)
	clusterLister := gke.NewFakeClusterLister(newListClusterResp(testutil.GetGKEClusters()), nil)
	nsListerDeleter := k8s.NewFakeNamespaceListerDeleter(&v1.NamespaceList{}, nil, nil)
	googleConfig := &config.Google{AccountFileJSON: "test", ProjectID: "proj1", Zone: "zone1"}
	hdl := DeleteLease(getterUpdater, "claimer", clusterLister, nil, nil, googleConfig, true, k8s.GetNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease/google/"+uuid.New(), nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusConflict, "response code")
}

func TestDeleteLeaseNoSuchLease(t *testing.T) {
	getterUpdater := k8s.NewFakeServiceGetterUpdater(
		&v1.Service{ObjectMeta: v1.ObjectMeta{Annotations: map[string]string{}}},
		nil,
		nil,
		nil,
	)
	clusterLister := gke.NewFakeClusterLister(newListClusterResp(testutil.GetGKEClusters()), nil)
	nsListerDeleter := k8s.NewFakeNamespaceListerDeleter(&v1.NamespaceList{}, nil, nil)
	googleConfig := &config.Google{AccountFileJSON: "test", ProjectID: "proj1", Zone: "zone1"}
	hdl := DeleteLease(getterUpdater, "claimer", clusterLister, nil, nil, googleConfig, true, k8s.GetNSFunc(nsListerDeleter, nil))
	req, err := http.NewRequest("DELETE", "/lease/google/"+uuid.New(), nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusConflict, "response code")
}

func TestDeleteLeaseExists(t *testing.T) {
	leaseableClusters := testutil.GetGKEClusters()
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
		getterUpdater := k8s.NewFakeServiceGetterUpdater(
			&v1.Service{ObjectMeta: v1.ObjectMeta{Annotations: annos}},
			nil,
			nil,
			nil,
		)
		path := "/lease/google/" + u.String()

		defaultNamespaces := []string{"default", "kube-system"}
		namespaces := append(defaultNamespaces, "ns1", "ns2")
		nsList := v1.NamespaceList{}
		for _, namespace := range namespaces {
			nsList.Items = append(nsList.Items, v1.Namespace{ObjectMeta: v1.ObjectMeta{Name: namespace}})
		}

		listClusterResp := newListClusterResp(testutil.GetGKEClusters())
		clusterLister := gke.NewFakeClusterLister(listClusterResp, nil)
		nsListerDeleter := k8s.NewFakeNamespaceListerDeleter(&nsList, nil, nil)
		googleConfig := &config.Google{AccountFileJSON: "test", ProjectID: "proj1", Zone: "zone1"}
		hdl := DeleteLease(getterUpdater, "claimer", clusterLister, nil, nil, googleConfig, true, k8s.GetNSFunc(nsListerDeleter, nil))
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
	leaseableClusters := testutil.GetGKEClusters()
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
		getterUpdater := k8s.NewFakeServiceGetterUpdater(
			&v1.Service{ObjectMeta: v1.ObjectMeta{Annotations: annos}},
			nil,
			nil,
			nil,
		)
		path := "/lease/google/" + u.String()

		defaultNamespaces := []string{"default", "kube-system"}
		namespaces := append(defaultNamespaces, "ns1", "ns2")
		nsList := v1.NamespaceList{}
		for _, namespace := range namespaces {
			nsList.Items = append(nsList.Items, v1.Namespace{ObjectMeta: v1.ObjectMeta{Name: namespace}})
		}

		listClusterResp := newListClusterResp(leaseableClusters)
		clusterLister := gke.NewFakeClusterLister(listClusterResp, nil)
		nsListerDeleter := k8s.NewFakeNamespaceListerDeleter(&nsList, nil, nil)
		googleConfig := &config.Google{AccountFileJSON: "test", ProjectID: "proj1", Zone: "zone1"}
		hdl := DeleteLease(getterUpdater, "claimer", clusterLister, nil, nil, googleConfig, false, k8s.GetNSFunc(nsListerDeleter, nil))
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
