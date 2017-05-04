package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	container "google.golang.org/api/container/v1"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/providers/gke"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
)

func TestWithAuthValidToken(t *testing.T) {
	cluster := &container.Cluster{
		Name:       "cluster1",
		Endpoint:   "192.168.1.1",
		MasterAuth: &container.MasterAuth{},
	}
	clusterLister := gke.NewFakeClusterLister(&container.ListClustersResponse{
		Clusters: []*container.Cluster{cluster},
	}, nil)
	services := k8s.NewFakeServiceGetterUpdater(&v1.Service{
		ObjectMeta: v1.ObjectMeta{Name: "service1"},
	}, nil, nil, nil)
	hdl := CreateLease(clusterLister, services, "", "", "")
	createLeaseHandler := htp.MethodMux(map[htp.Method]http.Handler{htp.Post: hdl})

	mux := http.NewServeMux()
	mux.Handle("/lease", WithAuth("auth token", "Authorization", createLeaseHandler))
	reqBody := `{"max_time":30, "cloud_provider":"google"}`
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
