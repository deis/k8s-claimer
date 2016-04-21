package handlers

import (
	"reflect"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	container "google.golang.org/api/container/v1"
)

// check for no clusters found
func TestSearchForFreeClusterNoneAvailable(t *testing.T) {
	leaseMap, err := leases.ParseMapFromAnnotations(map[string]string{})
	assert.NoErr(t, err)
	clusterLister := gke.FakeClusterLister{Err: nil, Resp: &container.ListClustersResponse{Clusters: nil}}
	clusterMap, err := clusters.ParseMapFromGKE(clusterLister, "", "")
	assert.NoErr(t, err)
	cluster, err := searchForFreeCluster(clusterMap, leaseMap)
	assert.Nil(t, cluster, "cluster")
	switch tErr := err.(type) {
	case errNoAvailableOrExpiredClustersFound:
	default:
		t.Fatalf("returned error was not a errNoAvailableOrExpiredClustersFound (it was a %s)", reflect.TypeOf(tErr))
	}
}

// check for expired lease but found no clusters
func TestSearchForClusterNoLeaseFound(t *testing.T) {
	const clusterName = "cluster1"
	leaseMap, err := leases.ParseMapFromAnnotations(map[string]string{
		uuid.New(): testutil.LeaseJSON(clusterName, time.Now().Add(-1*time.Hour), leases.TimeFormat),
	})
	assert.NoErr(t, err)
	clusterLister := gke.FakeClusterLister{
		Err:  nil,
		Resp: &container.ListClustersResponse{Clusters: nil},
	}
	clusterMap, err := clusters.ParseMapFromGKE(clusterLister, "", "")
	assert.NoErr(t, err)
	cluster, err := searchForFreeCluster(clusterMap, leaseMap)
	assert.Nil(t, cluster, "cluster")
	switch tErr := err.(type) {
	case errExpiredLeaseGKEMissing:
		assert.Equal(t, tErr.clusterName, clusterName, "cluster name")
	default:
		t.Fatalf("returned error was not a errExpiredLeaseGKEMissing (it was a %s)", reflect.TypeOf(tErr))
	}
}
