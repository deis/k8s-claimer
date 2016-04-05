package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arschles/assert"
	"github.com/pborman/uuid"
)

func TestCreateLeaseInvalidReq(t *testing.T) {
	hdl := CreateLease(nil)
	req, err := http.NewRequest("POST", "/lease", bytes.NewReader(nil))
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestCreateLeaseValidResp(t *testing.T) {
	hdl := CreateLease(nil)
	reqBody := `{"max_time":30}`
	req, err := http.NewRequest("POST", "/lease", strings.NewReader(reqBody))
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusOK, "response code")
	leaseResp := new(createLeaseResp)
	assert.NoErr(t, json.NewDecoder(res.Body).Decode(leaseResp))
	assert.Equal(t, leaseResp.KubeConfig, "", "returned kubeconfig")
	assert.Equal(t, leaseResp.IP, "", "returned IP address")
	parsedUUID := uuid.Parse(leaseResp.Token)
	assert.True(t, parsedUUID != nil, "returned token is not a valid uuid")
}
