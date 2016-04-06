package leases

import (
	"encoding/json"

	"github.com/pborman/uuid"
)

// Map holds an in-memory representation of the set of leases written to k8s annotations.
// It can look up leases by lease token (which is a UUID) or cluster name
type Map struct {
	// mapping from uuid to lease. this map is what's stored in the k8s annotation
	uuidMap map[string]*Lease
	// mapping from cluster name to uuid. this map is the secondary index into uuidMap
	nameMap map[string]uuid.UUID
}

// ParseMapFromAnnotations parses a map of Kubernetes annotations into a lease map. Returns nil
// and an appropriate error if the parsing failed, non-nil and no error otherwise
func ParseMapFromAnnotations(annotations map[string]string) (*Map, error) {
	uuidMap := make(map[string]*Lease)
	nameMap := make(map[string]uuid.UUID)
	for uuidStr, leaseStr := range annotations {
		u := uuid.Parse(uuidStr)
		if u == nil {
			return nil, ErrMalformedUUID{uuidStr: uuidStr}
		}
		lease, err := ParseLease(leaseStr)
		if err != nil {
			return nil, ErrParseLease{leaseStr: leaseStr, parseErr: err}
		}
		uuidMap[u.String()] = lease
		nameMap[lease.ClusterName] = u
	}
	return &Map{uuidMap: uuidMap, nameMap: nameMap}, nil
}

// LeaseByClusterName finds a lease in m by the given cluster name. returns nil and false if no
// lease exists for the given cluster name, non-nil and true otherwise
func (m Map) LeaseByClusterName(clusterName string) (*Lease, bool) {
	u, ok := m.nameMap[clusterName]
	if !ok {
		return nil, false
	}
	l, ok := m.uuidMap[u.String()]
	if !ok {
		return nil, false
	}
	return l, true
}

// UUIDs returns the map of lease UUIDs to the lease for that UUID
func (m Map) UUIDs() ([]uuid.UUID, error) {
	ret := make([]uuid.UUID, len(m.uuidMap))
	i := 0
	for uuidStr := range m.uuidMap {
		u := uuid.Parse(uuidStr)
		if u == nil {
			return nil, ErrMalformedUUID{uuidStr: uuidStr}
		}
		ret[i] = u
		i++
	}
	return ret, nil
}

// LeaseForUUID looks up a lease for the given UUID. Returns nil and false if none was found,
// non-nil and true otherwise
func (m Map) LeaseForUUID(u uuid.UUID) (*Lease, bool) {
	l, ok := m.uuidMap[u.String()]
	return l, ok
}

// CreateLease attempts to set the given lease under the given uuid. If u already existed or
// l.ClusterName otherwise already has a lease associated with it, does nothing and returns false.
// Otherwise adds the lease to the map and returns true
func (m *Map) CreateLease(u uuid.UUID, l *Lease) bool {
	if _, found := m.uuidMap[u.String()]; found {
		return false
	}
	m.uuidMap[u.String()] = l
	m.nameMap[l.ClusterName] = u
	return true
}

// DeleteLease attempts to delete the lease under the given uuid. If there is no such lease,
// does nothing and returns false. Otherwise, completes the delete operation and returns true
func (m *Map) DeleteLease(u uuid.UUID) bool {
	lease, found := m.uuidMap[u.String()]
	if !found {
		return false
	}
	delete(m.uuidMap, u.String())
	delete(m.nameMap, lease.ClusterName)
	return true
}

// ToAnnotations returns a raw map[string]string of lease tokens and json-encoded leases. This map is
// suitable for use in Kubernetes annotations, and will be parseable by ParseMapFromAnnotations
func (m *Map) ToAnnotations() (map[string]string, error) {
	ret := make(map[string]string)
	for token, lease := range m.uuidMap {
		leaseBytes, err := json.Marshal(lease)
		if err != nil {
			return map[string]string{}, err
		}
		ret[token] = string(leaseBytes)
	}
	return ret, nil
}
