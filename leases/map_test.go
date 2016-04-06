package leases

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
)

func TestParseMapFromAnnotations(t *testing.T) {
	rawAnnotations := testutil.GetRawAnnotations(
		testutil.GetClusterNames(),
		TimeFormat,
		testutil.DefaultTimeFunc,
		testutil.DefaultUUIDFunc,
	)
	m, err := ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)
	for name, uuid := range m.nameMap {
		lease, found := m.uuidMap[uuid.String()]
		assert.True(t, found, "lease %s not found in uuid map", name)
		assert.Equal(t, lease.ClusterName, name, "lease cluster name")
	}
}

func TestLeaseByClusterName(t *testing.T) {
	clusterNames := testutil.GetClusterNames()
	rawAnnotations := testutil.GetRawAnnotations(
		clusterNames,
		TimeFormat,
		testutil.DefaultTimeFunc,
		testutil.DefaultUUIDFunc,
	)
	m, err := ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)
	l, found := m.LeaseByClusterName("no such cluster")
	assert.True(t, l == nil, "lease returned when nil expected")
	assert.False(t, found, "found reported true when false expected")
	for _, clusterName := range clusterNames {
		l, found := m.LeaseByClusterName(clusterName)
		assert.Equal(t, l.ClusterName, clusterName, "cluster name")
		assert.True(t, found, "found reported false for lease %s, expected true", clusterName)
	}
}

func TestUUIDs(t *testing.T) {
	rawAnnotations := testutil.GetRawAnnotations(
		testutil.GetClusterNames(),
		TimeFormat,
		testutil.DefaultTimeFunc,
		testutil.DefaultUUIDFunc,
	)
	m, err := ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)
	uuids, err := m.UUIDs()
	assert.NoErr(t, err)
	assert.Equal(t, len(uuids), len(rawAnnotations), "number of returned UUIDs")
	for _, u := range uuids {
		_, found := rawAnnotations[u.String()]
		assert.True(t, found, "uuid %s not found in raw annotations", u.String())
	}
}

func TestCreateDeleteLease(t *testing.T) {
	clusterNames := testutil.GetClusterNames()
	rawAnnotations := testutil.GetRawAnnotations(
		clusterNames,
		TimeFormat,
		testutil.DefaultTimeFunc,
		testutil.DefaultUUIDFunc,
	)
	m, err := ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)

	newUUID := uuid.NewUUID()
	newLease := NewLease("cluster 12345", time.Now().Add(1*time.Hour))
	assert.True(t, m.CreateLease(newUUID, newLease), "failed to create a new lease")
	l, found := m.LeaseForUUID(newUUID)
	assert.Equal(t, l, newLease, "newly added lease")
	assert.True(t, found, "lease for cluster %s not found when fetched by uuid %s", newLease.ClusterName, newUUID)
	l, found = m.LeaseByClusterName(newLease.ClusterName)
	assert.Equal(t, l, newLease, "newly added lease")
	assert.True(t, found, "lease for cluster %s not found when fetched by cluster name", newLease.ClusterName)

	assert.False(t, m.CreateLease(newUUID, newLease), "was able to create a new, duplicate lease")

	newUUID2 := uuid.NewUUID()
	newLease2 := NewLease(clusterNames[0], time.Now().Add(1*time.Hour))
	assert.False(t, m.CreateLease(newUUID2, newLease2), "was able to create a new lease with duplicate cluster name")

	assert.True(t, m.DeleteLease(newUUID), "failed to delete existing lease")
	assert.False(t, m.DeleteLease(newUUID), "was able to delete a lease that doesn't exist")

	l, found = m.LeaseForUUID(newUUID)
	assert.True(t, l == nil, "lease for cluster %s was found by uuid %s after it was deleted", newLease.ClusterName, newUUID)
	assert.False(t, found, "lease for cluster %s was found by uuid %s after it was deleted", newLease.ClusterName, newUUID)
	l, found = m.LeaseByClusterName(newLease.ClusterName)
	assert.True(t, l == nil, "lease for cluster %s was found by name after it was deleted", newLease.ClusterName)
	assert.False(t, found, "lease for cluster %s was found by name after it was deleted", newLease.ClusterName)
}

func TestToAnnotations(t *testing.T) {
	m := &Map{
		uuidMap: map[string]*Lease{
			uuid.New(): &Lease{ClusterName: "cluster1"},
			uuid.New(): &Lease{ClusterName: "cluster2"},
		},
		// make nameMap empty - ToAnnotations should just look at uuidMap b/c that's the only
		// data that goes in to the annotation
		nameMap: map[string]uuid.UUID{},
	}
	annos, err := m.ToAnnotations()
	assert.NoErr(t, err)
	i := 0
	for u, leaseStr := range annos {
		lease, found := m.uuidMap[u]
		assert.True(t, found, "lease for uuid %s (%#%d) not found", u, i)
		leaseBytes, err := json.Marshal(lease)
		assert.NoErr(t, err)
		assert.Equal(t, leaseStr, string(leaseBytes), "lease string")
		i++
	}
}
