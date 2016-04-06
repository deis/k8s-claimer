package handlers

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	container "google.golang.org/api/container/v1"
	"k8s.io/kubernetes/pkg/api"
)

var (
	errUnusedGKEClusterNotFound = errors.New("unused GKE cluster not found")
	errNoExpiredLeases          = errors.New("no expired leases exist")
)

type leaseAnnotationValue struct {
	ClusterName         string `json:"cluster_name"`
	LeaseExpirationTime string `json:"lease_expiration_time"`
}

func getLeasesFromAnnotations(annotations map[string]string) map[string]*leaseAnnotationValue {
	ret := make(map[string]*leaseAnnotationValue)
	for clusterName, annoValStr := range annotations {
		annoVal := new(leaseAnnotationValue)
		if err := json.NewDecoder(strings.NewReader(annoValStr)).Decode(annoVal); err != nil {
			continue
		}
		ret[clusterName] = annoVal
	}
	return ret
}

// findUnusedGKECluster finds a GKE cluster that's not currently in use according to the
// annotations in svc. returns errUnusedGKEClusterNotFound if none is found
func findUnusedGKECluster(gkeClusters []*container.Cluster, svc *api.Service) (*container.Cluster, error) {
	existingLeases := getLeasesFromAnnotations(svc.Annotations)
	for _, cluster := range gkeClusters {
		_, found := existingLeases[cluster.Name]
		if !found {
			return cluster, nil
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

func findExpiredLease(svc *api.Service) (string, error) {
	leases := getLeasesFromAnnotations(svc.Annotations)
	now := time.Now()
	//...
	return "", nil
}
