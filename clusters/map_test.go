package clusters

import (
	"testing"

	"github.com/arschles/assert"
	container "google.golang.org/api/container/v1"
)

func TestClustersToMap(t *testing.T) {
	cl1 := &container.Cluster{Name: "cluster1"}
	cl2 := &container.Cluster{Name: "cluster2"}
	slice := []*container.Cluster{cl1, cl2}
	m := clustersToMap(slice)
	assert.Equal(t, m[cl1.Name], cl1, "cluster 1")
	assert.Equal(t, m[cl2.Name], cl2, "cluster 2")
}
