package handlers

import (
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	container "google.golang.org/api/container/v1"
)

const (
	projID = "test project"
	zone   = "test zone"
)

func TestFindUnusedGKECluster(t *testing.T) {
	clusterNames := testutil.GetClusterNames()

	// test when all clusters are leased
	fakeLister := &gke.FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: nil},
		Err:  nil,
	}
	for _, clusterName := range clusterNames {
		fakeLister.Resp.Clusters = append(fakeLister.Resp.Clusters, &container.Cluster{Name: clusterName})
	}

	clusterMap, err := clusters.ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)

	rawAnnotations := testutil.GetRawAnnotations(
		clusterNames,
		leases.TimeFormat,
		testutil.DefaultTimeFunc,
		testutil.DefaultUUIDFunc,
	)
	leaseMap, err := leases.ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)
	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap)
	assert.True(t, unusedCluster == nil, "unused cluster returned non-nil when all clusters were in use")
	assert.Err(t, errUnusedGKEClusterNotFound, err)

	// test when there is a cluster that's not leased
	var freedUUID uuid.UUID
	for uuidStr := range rawAnnotations {
		parsedUUID := uuid.Parse(uuidStr)
		assert.True(t, parsedUUID != nil, "uuid parsed from %s was invalid", uuidStr)
		freedUUID = parsedUUID
		break
	}

	freedLease, found := leaseMap.LeaseForUUID(freedUUID)
	assert.True(t, found, "lease for uuid %s not found", freedUUID)
	deleted := leaseMap.DeleteLease(freedUUID)
	assert.True(t, deleted, "lease for cluster %s was not deleted", freedLease.ClusterName)

	unusedCluster, err = findUnusedGKECluster(clusterMap, leaseMap)
	assert.NoErr(t, err)
	assert.Equal(t, unusedCluster.Name, freedLease.ClusterName, "free cluster name")
}
