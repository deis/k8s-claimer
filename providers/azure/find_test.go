package azure

import (
	"reflect"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/containerservice"
	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
)

var (
	clusterName = "cluster1"
	cluster1    = containerservice.ContainerService{ID: &clusterName, Name: &clusterName}
	azureConfig = config.Azure{ClientID: "clientID", ClientSecret: "clientSecret", TenantID: "tenantID", SubscriptionID: "subID"}
)

// check for no clusters found
func TestSearchForFreeClusterNoneAvailable(t *testing.T) {
	leaseMap, err := leases.ParseMapFromAnnotations(map[string]string{})
	assert.NoErr(t, err)
	clusterLister := FakeClusterLister{Err: nil, Resp: &containerservice.ListResult{Value: nil}}
	clusterMap, err := ParseMapFromAzure(clusterLister)
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
		Resp: &containerservice.ListResult{Value: nil},
	}
	clusterMap, err := ParseMapFromAzure(clusterLister)
	assert.NoErr(t, err)
	cluster, err := searchForFreeCluster(clusterMap, leaseMap, "", "")
	assert.Nil(t, cluster, "cluster")
	assert.Err(t, errNoAvailableOrExpiredClustersFound{}, err)
}

func TestFindUnusedAzureClusterByName(t *testing.T) {
	leaseableClusters := testutil.GetAzureClusters()

	// test when all clusters are leased
	fakeLister := &FakeClusterLister{
		Resp: &containerservice.ListResult{Value: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := ParseMapFromAzure(fakeLister)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedCluster(clusterMap, leaseMap, "getClusterByName", "")
	assert.NoErr(t, err)
	assert.Equal(t, *unusedCluster.Name, "getClusterByName", "free cluster name")
}

func TestFindRandomUnusedAzureCluster(t *testing.T) {
	leaseableClusters := testutil.GetAzureClusters()

	// test when all clusters are leased
	fakeLister := &FakeClusterLister{
		Resp: &containerservice.ListResult{Value: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := ParseMapFromAzure(fakeLister)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedCluster(clusterMap, leaseMap, "", "")
	assert.NoErr(t, err)
	assert.NotNil(t, unusedCluster, "free cluster name")
}

func TestGetClusterFromLease(t *testing.T) {
	clusterLister := FakeClusterLister{
		Resp: &containerservice.ListResult{Value: &[]containerservice.ContainerService{cluster1}},
	}
	lease := leases.NewLease(clusterName, time.Now().Add(1*time.Hour))
	cluster, err := GetClusterFromLease(lease, clusterLister)
	assert.NoErr(t, err)
	assert.Equal(t, cluster.Name, cluster1.Name, "cluster name")
}
