package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
)

func TestDeleteLeaseNoToken(t *testing.T) {
	getterUpdater := newFakeServiceGetterUpdater(nil, nil, nil, nil)
	hdl := DeleteLease(getterUpdater, "claimer")
	req, err := http.NewRequest("DELETE", "/lease", nil)
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidLeaseToken(t *testing.T) {
	getterUpdater := newFakeServiceGetterUpdater(nil, nil, nil, nil)
	hdl := DeleteLease(getterUpdater, "claimer")
	req, err := http.NewRequest("DELETE", "/lease/abcd", nil)
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestDeleteLeaseInvalidAnnotations(t *testing.T) {
	getterUpdater := newFakeServiceGetterUpdater(
		&api.Service{ObjectMeta: api.ObjectMeta{Annotations: map[string]string{"a": "b"}}},
		nil,
		nil,
		nil,
	)
	hdl := DeleteLease(getterUpdater, "claimer")
	req, err := http.NewRequest("DELETE", "/lease/"+uuid.New(), nil)
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusInternalServerError, "response code")
}

func TestDeleteLeaseNoSuchLease(t *testing.T) {
	getterUpdater := newFakeServiceGetterUpdater(
		&api.Service{ObjectMeta: api.ObjectMeta{Annotations: map[string]string{}}},
		nil,
		nil,
		nil,
	)
	hdl := DeleteLease(getterUpdater, "claimer")
	req, err := http.NewRequest("DELETE", "/lease/"+uuid.New(), nil)
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusConflict, "response code")
}

func TestDeleteLeaseExists(t *testing.T) {
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
		hdl := DeleteLease(getterUpdater, "claimer")
		req, err := http.NewRequest("DELETE", path, nil)
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
	}
}
