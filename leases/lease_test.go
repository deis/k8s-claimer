package leases

import (
	"fmt"
	"testing"
	"time"

	"github.com/arschles/assert"
)

const (
	clusterName = "my test cluster"
)

func TestParseLease(t *testing.T) {
	tme := time.Now()
	tStr := tme.Format(timeFormat)
	jsonStr := fmt.Sprintf(`{"cluster_name":"%s", "lease_expiration_time":"%s"}`, clusterName, tStr)
	lease, err := ParseLease(jsonStr)
	assert.NoErr(t, err)
	assert.Equal(t, lease.ClusterName, clusterName, "cluster name")
	assert.Equal(t, lease.LeaseExpirationTime, tStr, "expiration time string")
}

func TestNewLease(t *testing.T) {
	tme := time.Now()
	l := NewLease(clusterName, tme)
	assert.Equal(t, l.ClusterName, clusterName, "cluster name")
	assert.Equal(t, l.LeaseExpirationTime, tme.Format(timeFormat), "lease expiration time")
}

func TestExpirationTime(t *testing.T) {
	tme := time.Now()
	l := NewLease(clusterName, tme)
	ex, err := l.ExpirationTime()
	assert.NoErr(t, err)
	assert.Equal(t, ex.Format(timeFormat), tme.Format(timeFormat), "expiration time")

	l = &Lease{ClusterName: clusterName, LeaseExpirationTime: "invalid expiration time"}
	ex, err = l.ExpirationTime()
	assert.True(t, err != nil, "error wasn't returned when it should have been")
	assert.True(t, ex.IsZero(), "non-zero time was returned when it should have been zero")
}
