package leases

import (
	"encoding/json"
	"time"
)

const (
	timeFormat = time.RFC3339
)

var (
	zeroTime time.Time
)

type Lease struct {
	ClusterName         string `json:"cluster_name"`
	LeaseExpirationTime string `json:"lease_expiration_time"`
}

func NewLease(clusterName string, exprTime time.Time) *Lease {
	return &Lease{
		ClusterName:         clusterName,
		LeaseExpirationTime: exprTime.Format(timeFormat),
	}
}

func ParseLease(leaseStr string) (*Lease, error) {
	l := new(Lease)
	if err := json.Unmarshal([]byte(leaseStr), l); err != nil {
		return nil, err
	}
	return l, nil
}

func (l Lease) ExpirationTime() (time.Time, error) {
	t, err := time.Parse(timeFormat, l.LeaseExpirationTime)
	if err != nil {
		return zeroTime, err
	}
	return t, nil
}
