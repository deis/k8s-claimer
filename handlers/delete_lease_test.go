package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arschles/assert"
)

func TestDeleteLease(t *testing.T) {
	hdl := DeleteLease()
	req, err := http.NewRequest("DELETE", "/lease", nil)
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusOK, "response code")
}
