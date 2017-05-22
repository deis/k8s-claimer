package azure

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/containerservice"
	"github.com/arschles/assert"
)

func TestClusterByName(t *testing.T) {
	cluster1 := "cluster1"
	cluster2 := "cluster2"
	nameMap := map[string]*containerservice.ContainerService{
		"cluster1": &containerservice.ContainerService{Name: &cluster1},
		"cluster2": &containerservice.ContainerService{Name: &cluster2},
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
	nameMap := make(map[string]*containerservice.ContainerService)
	for _, name := range names {
		nameMap[name] = &containerservice.ContainerService{Name: &name}
	}
	m := &Map{nameMap: nameMap}
	retNames := m.Names()
	assert.Equal(t, len(retNames), len(names), "length of names slice")
}
