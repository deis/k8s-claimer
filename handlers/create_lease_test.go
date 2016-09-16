package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	container "google.golang.org/api/container/v1"
	"gopkg.in/yaml.v2"
	k8sapi "k8s.io/kubernetes/pkg/api"
)

var (
	expectedCluster = &container.Cluster{
		Name:               "cluster1",
		CurrentNodeVersion: "9.9.9",
		Endpoint:           "192.168.1.1",
		MasterAuth:         &container.MasterAuth{},
	}

	expectedListClusterResp = newListClusterResp([]*container.Cluster{expectedCluster})
)

func newListClusterResp(clusters []*container.Cluster) *container.ListClustersResponse {
	resp := &container.ListClustersResponse{Clusters: make([]*container.Cluster, len(clusters))}
	for i, cluster := range clusters {
		resp.Clusters[i] = cluster
	}
	return resp
}

func newFakeClusterLister(resp *container.ListClustersResponse, err error) *gke.FakeClusterLister {
	return &gke.FakeClusterLister{
		Resp: resp,
		Err:  err,
	}
}

func TestCreateLeaseInvalidReq(t *testing.T) {
	cl := newFakeClusterLister(nil, nil)
	slu := newFakeServiceGetterUpdater(nil, nil, nil, nil)
	hdl := CreateLease(cl, slu, "", "", "")
	req, err := http.NewRequest("POST", "/lease", bytes.NewReader(nil))
	assert.NoErr(t, err)
	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadRequest, "response code")
}

func TestCreateLeaseValidResp(t *testing.T) {
	cluster := testutil.GetClusters()[0]
	clusterLister := newFakeClusterLister(newListClusterResp([]*container.Cluster{cluster}), nil)
	services := newFakeServiceGetterUpdater(&k8sapi.Service{
		ObjectMeta: k8sapi.ObjectMeta{Name: "service1"},
	}, nil, nil, nil)
	hdl := CreateLease(clusterLister, services, "", "", "")
	reqBody := `{"max_time":30}`
	req, err := http.NewRequest("POST", "/lease", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "some awesome token")
	assert.NoErr(t, err)

	res := httptest.NewRecorder()
	hdl.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusOK, "response code")

	leaseResp := new(api.CreateLeaseResp)
	assert.NoErr(t, json.NewDecoder(res.Body).Decode(leaseResp))

	expectedKubeCfg, err := createKubeConfigFromCluster(expectedCluster)
	assert.NoErr(t, err)

	expectedMarshalledKubeCfg, err := yaml.Marshal(expectedKubeCfg)
	assert.NoErr(t, err)
	marshalledBytes, err := leaseResp.KubeConfigBytes()
	assert.NoErr(t, err)
	assert.Equal(t, marshalledBytes, expectedMarshalledKubeCfg, "returned kubeconfig")
	assert.Equal(t, leaseResp.IP, cluster.Endpoint, "returned IP address")

	parsedUUID := uuid.Parse(leaseResp.Token)
	assert.True(t, parsedUUID != nil, "returned token is not a valid uuid")
}
