package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arschles/assert"
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
