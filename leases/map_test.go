package leases

import (
	"encoding/json"
	"testing"

	"github.com/arschles/assert"
	"github.com/pborman/uuid"
)

func TestToAnnotations(t *testing.T) {
	m := &Map{
		uuidMap: map[string]*Lease{
			uuid.New(): &Lease{ClusterName: "cluster1"},
			uuid.New(): &Lease{ClusterName: "cluster2"},
		},
		// make nameMap empty - ToAnnotations should just look at uuidMap b/c that's the only
		// data that goes in to the annotation
		nameMap: map[string]uuid.UUID{},
	}
	annos, err := m.ToAnnotations()
	assert.NoErr(t, err)
	i := 0
	for u, leaseStr := range annos {
		lease, found := m.uuidMap[u]
		assert.True(t, found, "lease for uuid %s (%#%d) not found", u, i)
		leaseBytes, err := json.Marshal(lease)
		assert.NoErr(t, err)
		assert.Equal(t, leaseStr, string(leaseBytes), "lease string")
		i++
	}
}
