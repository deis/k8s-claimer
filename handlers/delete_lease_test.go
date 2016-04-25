package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
)

func TestDeleteLeaseNoToken(t *testing.T) {
	os.Setenv("AUTH_TOKEN", "some awesome token")
	defer os.Clearenv()
	getterUpdater := newFakeServiceGetterUpdater(nil, nil, nil, nil)
	nsListerDeleter := newFakeNamespaceListerDeleter(nil, nil, nil)
	hdl := DeleteLease(getterUpdater, "claimer", nsListerDeleter)
	req, err := http.NewRequest("DELETE", "/lease", nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidLeaseToken(t *testing.T) {
	os.Setenv("AUTH_TOKEN", "some awesome token")
	defer os.Clearenv()
	getterUpdater := newFakeServiceGetterUpdater(nil, nil, nil, nil)
	nsListerDeleter := newFakeNamespaceListerDeleter(nil, nil, nil)
	hdl := DeleteLease(getterUpdater, "claimer", nsListerDeleter)
	req, err := http.NewRequest("DELETE", "/lease/abcd", nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidAnnotations(t *testing.T) {
	os.Setenv("AUTH_TOKEN", "some awesome token")
	defer os.Clearenv()
	// Issue a DELETE with a lease token that doesn't point to a valid lease.
	// Annotations have invalid data in them, which the lease parser should just ignore.
	getterUpdater := newFakeServiceGetterUpdater(
		&api.Service{ObjectMeta: api.ObjectMeta{Annotations: map[string]string{"a": "b"}}},
		nil,
		nil,
		nil,
	)
	nsListerDeleter := newFakeNamespaceListerDeleter(nil, nil, nil)
	hdl := DeleteLease(getterUpdater, "claimer", nsListerDeleter)
	req, err := http.NewRequest("DELETE", "/lease/"+uuid.New(), nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusConflict, "response code")
}

func TestDeleteLeaseNoSuchLease(t *testing.T) {
	os.Setenv("AUTH_TOKEN", "some awesome token")
	defer os.Clearenv()
	getterUpdater := newFakeServiceGetterUpdater(
		&api.Service{ObjectMeta: api.ObjectMeta{Annotations: map[string]string{}}},
		nil,
		nil,
		nil,
	)
	nsListerDeleter := newFakeNamespaceListerDeleter(nil, nil, nil)
	hdl := DeleteLease(getterUpdater, "claimer", nsListerDeleter)
	req, err := http.NewRequest("DELETE", "/lease/"+uuid.New(), nil)
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusConflict, "response code")
}

func TestDeleteLeaseInvalidAuth(t *testing.T) {
	os.Setenv("AUTH_TOKEN", "some awesome token")
	defer os.Clearenv()
	clusterNames := []string{"cluster1"}
	uuids := make([]uuid.UUID, len(clusterNames))
	annos := testutil.GetRawAnnotations(
		clusterNames,
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
	getterUpdater := newFakeServiceGetterUpdater(
		&api.Service{ObjectMeta: api.ObjectMeta{Annotations: annos}},
		nil,
		nil,
		nil,
	)
	path := "/lease/" + uuids[0].String()
	namespaces := []string{"ns1", "ns2"}
	nsList := api.NamespaceList{}
	for _, namespace := range namespaces {
		nsList.Items = append(nsList.Items, api.Namespace{ObjectMeta: api.ObjectMeta{Name: namespace}})
	}
	nsListerDeleter := newFakeNamespaceListerDeleter(
		&nsList, nil, nil)
	hdl := DeleteLease(getterUpdater, "claimer", nsListerDeleter)
	req, err := http.NewRequest("DELETE", path, nil)
	req.Header.Set("Authorization", "different token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusUnauthorized, "response code")
}

func TestDeleteLeaseExists(t *testing.T) {
	os.Setenv("AUTH_TOKEN", "some awesome token")
	defer os.Clearenv()
	clusterNames := []string{"cluster1", "cluster2"}
	uuids := make([]uuid.UUID, len(clusterNames))
	annos := testutil.GetRawAnnotations(
		clusterNames,
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
		nsListerDeleter := newFakeNamespaceListerDeleter(
			&nsList, nil, nil)
		hdl := DeleteLease(getterUpdater, "claimer", nsListerDeleter)
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
		assert.Equal(t, nsListerDeleter.NsDeleted, []string{"ns1", "ns2"}, "namespaces deleted")
	}
}
