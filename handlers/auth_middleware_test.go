package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/api/container/v1"

	k8sapi "k8s.io/kubernetes/pkg/api"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/htp"
)

func TestWithAuthValidToken(t *testing.T) {
	cluster := &container.Cluster{
		Name:       "cluster1",
		Endpoint:   "192.168.1.1",
		MasterAuth: &container.MasterAuth{},
	}
	clusterLister := newFakeClusterLister(&container.ListClustersResponse{
		Clusters: []*container.Cluster{cluster},
	}, nil)
	services := newFakeServiceGetterUpdater(&k8sapi.Service{
		ObjectMeta: k8sapi.ObjectMeta{Name: "service1"},
	}, nil, nil, nil)
	hdl := CreateLease(clusterLister, services, "", "", "")
	createLeaseHandler := htp.MethodMux(map[htp.Method]http.Handler{htp.Post: hdl})

	mux := http.NewServeMux()
	mux.Handle("/lease", WithAuth("auth token", "Authorization", createLeaseHandler))
	reqBody := `{"max_time":30}`
	req, err := http.NewRequest("POST", "/lease", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "auth token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusOK, "response code")
}

func TestWithAuthInvalidToken(t *testing.T) {
	hdl := CreateLease(nil, nil, "", "", "")
	createLeaseHandler := htp.MethodMux(map[htp.Method]http.Handler{htp.Post: hdl})

	mux := http.NewServeMux()
	mux.Handle("/lease", WithAuth("auth token", "Authorization", createLeaseHandler))
	req, err := http.NewRequest("POST", "/lease", bytes.NewReader(nil))
	req.Header.Set("Authorization", "invalid auth token")
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusUnauthorized, "response code")
}
