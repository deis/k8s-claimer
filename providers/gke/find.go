package gke

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	container "google.golang.org/api/container/v1"

	"github.com/deis/k8s-claimer/leases"
)

type errNoAvailableOrExpiredClustersFound struct{}

func (e errNoAvailableOrExpiredClustersFound) Error() string {
	return "no available or expired clusters found"
}

type errExpiredLeaseGKEMissing struct {
	clusterName string
}

var (
	errNoExpiredLeases = errors.New("no expired leases exist")
)

func (e errExpiredLeaseGKEMissing) Error() string {
	return fmt.Sprintf("cluster %s has an expired lease but does not exist in GKE", e.clusterName)
}

// searchForFreeCluster looks for an available GKE cluster to lease.
// It will try and match clusterRegex if possible or clusterVersion
//
// Returns errNoAvailableOrExpiredClustersFound if it found no free or expired lease
// Returns errExpiredLeaseGKEMissing if it found an expired lease but the cluster associated with
// that lease doesn't exist in GKE
func searchForFreeCluster(clusterMap *Map, leaseMap *leases.Map, clusterRegex string, clusterVersion string) (*container.Cluster, error) {
	uuidAndLeases, expiredLeaseErr := findExpiredLeases(leaseMap)
	if expiredLeaseErr == nil {
		for _, expiredLease := range uuidAndLeases {
			leaseMap.DeleteLease(expiredLease.UUID)
		}
	}
	cluster, err := findUnusedGKECluster(clusterMap, leaseMap, clusterRegex, clusterVersion)
	if err != nil {
		return nil, errNoAvailableOrExpiredClustersFound{}
	}
	return cluster, nil
}

// findExpiredLeases searches in the leases in the svc annotations and returns the cluster name of
// the first expired lease it finds. If none found, returns an empty string and errNoExpiredLeases
func findExpiredLeases(leaseMap *leases.Map) ([]*leases.UUIDAndLease, error) {
	var leasesUUIDs []*leases.UUIDAndLease
	now := time.Now()
	uuids, err := leaseMap.UUIDs()
	if err != nil {
		return nil, err
	}
	for _, u := range uuids {
		lease, _ := leaseMap.LeaseForUUID(u)
		exprTime, err := lease.ExpirationTime()
		if err != nil {
			return nil, err
		}
		if now.After(exprTime) {
			leasesUUIDs = append(leasesUUIDs, leases.NewUUIDAndLease(u, lease))
		}
	}
	if len(leasesUUIDs) > 0 {
		return leasesUUIDs, nil
	}
	return nil, errNoExpiredLeases
}

// findUnusedGKECluster finds a GKE cluster that's not currently in use according to the
// annotations in svc. It will also attempt to match the clusterRegex passed in if possible.
// Returns errUnusedGKEClusterNotFound if none is found
func findUnusedGKECluster(clusterMap *Map, leaseMap *leases.Map, clusterRegex string, clusterVersion string) (*container.Cluster, error) {
	if clusterRegex != "" {
		return findUnusuedGKEClusterByName(clusterMap, leaseMap, clusterRegex)
	} else if clusterVersion != "" {
		return findUnusedGKEClusterByVersion(clusterMap, leaseMap, clusterVersion)
	} else {
		return findRandomUnusuedGKECluster(clusterMap, leaseMap)
	}
}

// findUnusuedGKEClusterByName attempts to find a unused GKE cluster that matches the regex passed in via the cli.
func findUnusuedGKEClusterByName(clusterMap *Map, leaseMap *leases.Map, clusterRegex string) (*container.Cluster, error) {
	regex, err := regexp.Compile(clusterRegex)
	if err != nil {
		return nil, err
	}
	for _, clusterName := range clusterMap.Names() {
		if regex.MatchString(clusterName) {
			cluster, err := checkLease(clusterMap, leaseMap, clusterName)
			if err == nil {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

// findUnusedGKEClusterByVersion attempts to find an unused GKE cluster that matches the version passed in via the CLI.
func findUnusedGKEClusterByVersion(clusterMap *Map, leaseMap *leases.Map, clusterVersion string) (*container.Cluster, error) {
	clusterNames := clusterMap.ClusterNamesByVersion(clusterVersion)
	if len(clusterNames) > 0 {
		for tries := 0; tries < 10; tries++ {
			clusterName := clusterNames[rand.Intn(len(clusterNames))]
			cluster, err := checkLease(clusterMap, leaseMap, clusterName)
			if err == nil {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

// findUnusuedGKECluster attempts to find a random unused GKE cluster
func findRandomUnusuedGKECluster(clusterMap *Map, leaseMap *leases.Map) (*container.Cluster, error) {
	clusterNames := clusterMap.Names()
	if len(clusterNames) > 0 {
		for tries := 0; tries < 10; tries++ {
			clusterName := clusterNames[rand.Intn(len(clusterNames))]
			cluster, err := checkLease(clusterMap, leaseMap, clusterName)
			if err == nil {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

// checkLease takes a clusterName and determines if there is an available lease
func checkLease(clusterMap *Map, leaseMap *leases.Map, clusterName string) (*container.Cluster, error) {
	cluster, _ := clusterMap.ClusterByName(clusterName)
	_, isLeased := leaseMap.LeaseByClusterName(clusterName)
	if !isLeased {
		return cluster, nil
	}
	return nil, errUnusedGKEClusterNotFound
}

type errNoSuchCluster struct {
	name string
}

func (e errNoSuchCluster) Error() string {
	return fmt.Sprintf("no such cluster %s", e.name)
}

// GetClusterFromLease takes a lease and will find the appropriate cluster
func GetClusterFromLease(lease *leases.Lease, clusterLister ClusterLister, projID, zone string) (*container.Cluster, error) {
	clusterMap, err := ParseMapFromGKE(clusterLister, projID, zone)
	if err != nil {
		return nil, err
	}
	cl, exists := clusterMap.ClusterByName(lease.ClusterName)
	if !exists {
		return nil, errNoSuchCluster{name: lease.ClusterName}
	}

	return cl, nil
}
