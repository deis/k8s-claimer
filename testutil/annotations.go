package testutil

import (
	"time"

	"github.com/pborman/uuid"
)

// DefaultTimeFunc can be passed to GetRawAnnotations for the timeFunc parameter
func DefaultTimeFunc(i int) time.Time {
	return time.Now().Add(time.Duration(i) * time.Second)
}

// DefaultUUIDFunc can be passed to GetRawAnnotations for the uuidFunc parameter
func DefaultUUIDFunc(i int) uuid.UUID {
	return uuid.NewUUID()
}

// GetRawAnnotations constructs a map of raw annotations, each of which represents a lease for one
// of the clusters in clusterNAmes
func GetRawAnnotations(
	clusterNames []string,
	timeFmt string,
	timeFunc func(int) time.Time,
	uuidFunc func(int) uuid.UUID,
) map[string]string {

	ret := make(map[string]string)
	i := 0
	for _, clusterName := range clusterNames {
		ret[uuid.New()] = LeaseJSON(clusterName, timeFunc(i), timeFmt)
		i++
	}
	return ret
}
