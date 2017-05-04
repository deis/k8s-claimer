package gke

import (
	"reflect"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	container "google.golang.org/api/container/v1"
)

var (
	cluster1 = container.Cluster{Name: "cluster1"}
)

// check for no clusters found
func TestSearchForFreeClusterNoneAvailable(t *testing.T) {
	leaseMap, err := leases.ParseMapFromAnnotations(map[string]string{})
	assert.NoErr(t, err)
	clusterLister := FakeClusterLister{Err: nil, Resp: &container.ListClustersResponse{Clusters: nil}}
	clusterMap, err := ParseMapFromGKE(clusterLister, "", "")
	assert.NoErr(t, err)
	cluster, err := searchForFreeCluster(clusterMap, leaseMap, "", "")
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
	clusterLister := FakeClusterLister{
		Err:  nil,
		Resp: &container.ListClustersResponse{Clusters: nil},
	}
	clusterMap, err := ParseMapFromGKE(clusterLister, "", "")
	assert.NoErr(t, err)
	cluster, err := searchForFreeCluster(clusterMap, leaseMap, "", "")
	assert.Nil(t, cluster, "cluster")
	assert.Err(t, errNoAvailableOrExpiredClustersFound{}, err)
}

const (
	projID = "test project"
	zone   = "test zone"
)

func TestFindUnusedGKEClusterByName(t *testing.T) {
	leaseableClusters := testutil.GetClusters()

	// test when all clusters are leased
	fakeLister := &FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap, "getClusterByName", "")
	assert.NoErr(t, err)
	assert.Equal(t, unusedCluster.Name, "getClusterByName", "free cluster name")
}

func TestFindUnusedGKEClusterByVersion(t *testing.T) {
	leaseableClusters := testutil.GetClusters()

	// test when all clusters are leased
	fakeLister := &FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap, "", "1.1.1")
	assert.NoErr(t, err)
	assert.Equal(t, unusedCluster.Name, "getClusterByVersion", "free cluster name")
	assert.Equal(t, unusedCluster.CurrentNodeVersion, "1.1.1", "free cluster version")
}

func TestFindRandomUnusedGKECluster(t *testing.T) {
	leaseableClusters := testutil.GetClusters()

	// test when all clusters are leased
	fakeLister := &FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap, "", "")
	assert.NoErr(t, err)
	assert.NotNil(t, unusedCluster, "free cluster name")
}

func TestGetClusterFromLease(t *testing.T) {
	clusterLister := FakeClusterLister{
		Resp: &container.ListClustersResponse{
			Clusters: []*container.Cluster{&cluster1},
		},
	}
	lease := leases.NewLease(cluster1.Name, time.Now().Add(1*time.Hour))
	cluster, err := GetClusterFromLease(lease, clusterLister, projID, zone)
	assert.NoErr(t, err)
	assert.Equal(t, cluster.Name, cluster1.Name, "cluster name")
}
