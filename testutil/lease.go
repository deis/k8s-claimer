package testutil

import (
	"fmt"
	"time"
)

// LeaseJSON returns the json representative of what a lease should look like when encoded.
// It does so without first creating a *leases.Lease, so it should be used in tests to verify
// the lease wire representation
func LeaseJSON(clusterName string, exprTime time.Time, timeFmt string) string {
	return fmt.Sprintf(`{"cluster_name":"%s", "lease_expiration_time":"%s"}`, clusterName, exprTime.Format(timeFmt))
}
