package handlers

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	timeFormat = time.RFC3339
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

// findExpiredLease searches in the leases in the svc annotations and returns the cluster name of
// the first expired lease it finds. If none found, returns an empty string and errNoExpiredLeases
func findExpiredLease(annotations map[string]string) (string, error) {
	leases := getLeasesFromAnnotations(annotations)
	now := time.Now()
	for _, leaseInfo := range leases {
		exprTime, err := time.Parse(timeFormat, leaseInfo.LeaseExpirationTime)
		if err != nil {
			return "", err
		}
		if exprTime.Before(now) {
			return leaseInfo.ClusterName, nil
		}
	}
	return "", errNoExpiredLeases
}
