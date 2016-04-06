package testutil

import (
	"time"

	"github.com/pborman/uuid"
)

// GetRawAnnotations constructs a map of raw annotations, each of which represents a lease for one
// of the clusters in clusterNAmes
func GetRawAnnotations(clusterNames []string, timeFmt string) map[string]string {
	ret := make(map[string]string)
	i := 0
	for _, clusterName := range clusterNames {
		ret[uuid.New()] = LeaseJSON(clusterName, time.Now().Add(time.Duration(i)*time.Second), timeFmt)
		i++
	}
	return ret
}
