package leases

import (
	"encoding/json"
	"time"
)

const (
	TimeFormat = time.RFC3339
)

var (
	zeroTime time.Time
)

// Lease is the json-encodable struct that represents what's in the value of one lease annotation
// in k8s
type Lease struct {
	ClusterName         string `json:"cluster_name"`
	LeaseExpirationTime string `json:"lease_expiration_time"`
}

// NewLease creates a new lease with the given cluster name and expiration time
func NewLease(clusterName string, exprTime time.Time) *Lease {
	return &Lease{
		ClusterName:         clusterName,
		LeaseExpirationTime: exprTime.Format(TimeFormat),
	}
}

// ParseLease decodes leaseStr from json into a Lease structure. Returns nil and any decoding error
// if there was one, and a valid lease and nil otherwise
func ParseLease(leaseStr string) (*Lease, error) {
	l := new(Lease)
	if err := json.Unmarshal([]byte(leaseStr), l); err != nil {
		return nil, err
	}
	return l, nil
}

// ExpirationTime returns the expiration time of this lease, if l.LeaseExpirationTime was a well
// formed time string, returns the time and nil. Otherwise returns the zero value of time
// (i.e. t.IsZero() will return true) and a non-nil error
func (l Lease) ExpirationTime() (time.Time, error) {
	t, err := time.Parse(TimeFormat, l.LeaseExpirationTime)
	if err != nil {
		return zeroTime, err
	}
	return t, nil
}
