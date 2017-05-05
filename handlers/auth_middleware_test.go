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
	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/providers/gke"
)

func TestWithAuthValidToken(t *testing.T) {
	cluster := &container.Cluster{
		Name:       "cluster1",
		Endpoint:   "192.168.1.1",
		MasterAuth: &container.MasterAuth{},
	}
	gkeClusterLister := gke.NewFakeClusterLister(&container.ListClustersResponse{
		Clusters: []*container.Cluster{cluster},
	}, nil)
	services := k8s.NewFakeServiceGetterUpdater(&v1.Service{
		ObjectMeta: v1.ObjectMeta{Name: "service1"},
	}, nil, nil, nil)
	googleConfig := &config.Google{ProjectID: "proj1", Zone: "zone1"}
	hdl := CreateLease(services, "", gkeClusterLister, nil, nil, googleConfig)
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
	hdl := CreateLease(nil, "", nil, nil, nil, nil)
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
