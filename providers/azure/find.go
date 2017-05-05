package azure

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/containerservice"
	"github.com/deis/k8s-claimer/leases"
)

type errNoAvailableOrExpiredClustersFound struct{}

func (e errNoAvailableOrExpiredClustersFound) Error() string {
	return "no available or expired clusters found"
}

type errExpiredLeaseAzureMissing struct {
	clusterName string
}

var (
	errNoExpiredLeases            = errors.New("no expired leases exist")
	errUnusedAzureClusterNotFound = errors.New("all Azure clusters are in use")
)

func (e errExpiredLeaseAzureMissing) Error() string {
	return fmt.Sprintf("cluster %s has an expired lease but does not exist in Azure", e.clusterName)
}

// searchForFreeCluster looks for an available Azure cluster to lease.
// It will try and match clusterRegex if possible or clusterVersion
//
// Returns errNoAvailableOrExpiredClustersFound if it found no free or expired lease
// Returns errExpiredLeaseAzureMissing if it found an expired lease but the cluster associated with
// that lease doesn't exist in Azure
func searchForFreeCluster(clusterMap *Map, leaseMap *leases.Map, clusterRegex string, clusterVersion string) (*containerservice.ContainerService, error) {
	uuidAndLeases, expiredLeaseErr := findExpiredLeases(leaseMap)
	if expiredLeaseErr == nil {
		for _, expiredLease := range uuidAndLeases {
			leaseMap.DeleteLease(expiredLease.UUID)
		}
	}
	cluster, err := findUnusedCluster(clusterMap, leaseMap, clusterRegex, clusterVersion)
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

// findUnusedCluster finds a Azure cluster that's not currently in use according to the
// annotations in svc. It will also attempt to match the clusterRegex passed in if possible.
// Returns errUnusedClusterNotFound if none is found
func findUnusedCluster(clusterMap *Map, leaseMap *leases.Map, clusterRegex string, clusterVersion string) (*containerservice.ContainerService, error) {
	if clusterRegex != "" {
		return findUnusuedClusterByName(clusterMap, leaseMap, clusterRegex)
	} else if clusterVersion != "" {
		return findUnusedClusterByVersion(clusterMap, leaseMap, clusterVersion)
	} else {
		return findRandomUnusuedCluster(clusterMap, leaseMap)
	}
}

// findUnusuedClusterByName attempts to find a unused Azure cluster that matches the regex passed in via the cli.
func findUnusuedClusterByName(clusterMap *Map, leaseMap *leases.Map, clusterRegex string) (*containerservice.ContainerService, error) {
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
	return nil, errUnusedAzureClusterNotFound
}

// findUnusedClusterByVersion attempts to find an unused Azure cluster that matches the version passed in via the CLI.
func findUnusedClusterByVersion(clusterMap *Map, leaseMap *leases.Map, clusterVersion string) (*containerservice.ContainerService, error) {
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
	return nil, errUnusedAzureClusterNotFound
}

// findUnusuedCluster attempts to find a random unused Azure cluster
func findRandomUnusuedCluster(clusterMap *Map, leaseMap *leases.Map) (*containerservice.ContainerService, error) {
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
	return nil, errUnusedAzureClusterNotFound
}

// checkLease takes a clusterName and determines if there is an available lease
func checkLease(clusterMap *Map, leaseMap *leases.Map, clusterName string) (*containerservice.ContainerService, error) {
	cluster, _ := clusterMap.ClusterByName(clusterName)
	_, isLeased := leaseMap.LeaseByClusterName(clusterName)
	if !isLeased {
		return cluster, nil
	}
	return nil, errUnusedAzureClusterNotFound
}

type errNoSuchCluster struct {
	name string
}

func (e errNoSuchCluster) Error() string {
	return fmt.Sprintf("no such cluster %s", e.name)
}

// GetClusterFromLease takes a lease and will find the appropriate cluster
func GetClusterFromLease(lease *leases.Lease, clusterLister ClusterLister) (*containerservice.ContainerService, error) {
	clusterMap, err := ParseMapFromAzure(clusterLister)
	if err != nil {
		return nil, err
	}
	cl, exists := clusterMap.ClusterByName(lease.ClusterName)
	if !exists {
		return nil, errNoSuchCluster{name: lease.ClusterName}
	}

	return cl, nil
}
