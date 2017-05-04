package gke

import (
	"fmt"
	"testing"

	"github.com/arschles/assert"
	container "google.golang.org/api/container/v1"
)

func TestClusterByName(t *testing.T) {
	nameMap := map[string]*container.Cluster{
		"cluster1": &container.Cluster{Name: "cluster2"},
		"cluster2": &container.Cluster{Name: "cluster1"},
	}
	m := &Map{nameMap: nameMap}
	for expectedName, expectedCluster := range nameMap {
		retCluster, found := m.ClusterByName(expectedName)
		assert.True(t, found, "cluster %s not found", expectedName)
		assert.Equal(t, retCluster, expectedCluster, fmt.Sprintf("cluster %s", expectedName))
	}
}

func TestNames(t *testing.T) {
	names := []string{"1", "2", "3"}
	nameMap := make(map[string]*container.Cluster)
	for _, name := range names {
		nameMap[name] = &container.Cluster{Name: name}
	}
	m := &Map{nameMap: nameMap}
	retNames := m.Names()
	assert.Equal(t, len(retNames), len(names), "length of names slice")
}
