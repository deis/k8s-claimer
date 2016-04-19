package handlers

import (
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
)

func newFakeServiceGetter(svc *api.Service, err error) *k8s.FakeServiceGetter {
	return &k8s.FakeServiceGetter{Svc: svc, Err: err}
}

func newFakeServiceUpdater(retSvc *api.Service, err error) *k8s.FakeServiceUpdater {
	return &k8s.FakeServiceUpdater{RetSvc: retSvc, Err: err}
}

func newFakeServiceGetterUpdater(
	getSvc *api.Service,
	getErr error,
	updateSvc *api.Service,
	updateErr error,
) *k8s.FakeServiceGetterUpdater {
	return &k8s.FakeServiceGetterUpdater{
		FakeServiceGetter:  newFakeServiceGetter(getSvc, getErr),
		FakeServiceUpdater: newFakeServiceUpdater(updateSvc, updateErr),
	}
}

func newFakeNamespaceLister(nsList *api.NamespaceList, err error) *k8s.FakeNamespaceLister {
	return &k8s.FakeNamespaceLister{NsList: nsList, Err: err}
}

func newFakeNamespaceDeleter(err error) *k8s.FakeNamespaceDeleter {
	return &k8s.FakeNamespaceDeleter{Err: err}
}

func newFakeNamespaceListerDeleter(
	listNs *api.NamespaceList,
	listErr error,
	deleteErr error,
) *k8s.FakeNamespaceListerDeleter {
	return &k8s.FakeNamespaceListerDeleter{
		FakeNamespaceLister:  newFakeNamespaceLister(listNs, listErr),
		FakeNamespaceDeleter: newFakeNamespaceDeleter(deleteErr),
	}
}

func TestFindExpiredLease(t *testing.T) {
	clusterNames := testutil.GetClusterNames()
	expClusterIdx := -1
	uuids := make([]uuid.UUID, len(clusterNames))
	rawAnnotations := testutil.GetRawAnnotations(
		clusterNames,
		leases.TimeFormat,
		func(i int) time.Time {
			if expClusterIdx == -1 {
				expClusterIdx = i
				return time.Now().Add(-1 * time.Second)
			}
			return time.Now().Add(1 * time.Hour)
		},
		func(i int) uuid.UUID {
			ret := uuid.NewUUID()
			uuids[i] = ret
			return ret
		},
	)
	expClusterName := clusterNames[expClusterIdx]
	expClusterUUID := uuids[expClusterIdx]
	leaseMap, err := leases.ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)
	expLease, found := leaseMap.LeaseByClusterName(expClusterName)
	assert.True(t, found, "expired lease for cluster %s not found", expClusterName)
	expired, err := findExpiredLease(leaseMap)
	assert.NoErr(t, err)
	assert.Equal(t, expired.Lease.ClusterName, expLease.ClusterName, "cluster name")
	assert.Equal(t, expired.UUID.String(), expClusterUUID.String(), "lease UUID")

	rawAnnotations = testutil.GetRawAnnotations(
		clusterNames,
		leases.TimeFormat,
		func(int) time.Time { return time.Now().Add(1 * time.Hour) },
		testutil.DefaultUUIDFunc,
	)
	leaseMap, err = leases.ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)
	expired, err = findExpiredLease(leaseMap)
	assert.True(t, expired == nil, "non-nil expired lease returned when non were expired")
	assert.Err(t, errNoExpiredLeases, err)
}
