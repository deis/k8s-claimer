package azure

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/arm/containerservice"
)

// Map is a map from cluster name to ACS cluster
type Map struct {
	nameMap map[string]*containerservice.ContainerService
}

func clusterNamesToMap(c *[]containerservice.ContainerService) map[string]*containerservice.ContainerService {
	ret := make(map[string]*containerservice.ContainerService)
	if c == nil {
		return nil
	}
	for _, cluster := range *c {
		clusterCopy := cluster
		ret[*cluster.Name] = &clusterCopy
	}
	return ret
}

// ParseMapFromAzure calls the Azure API to get a list of clusters, then returns a Map representation
// of those clusters. Returns nil and an appropriate error if any errors occurred along the way
func ParseMapFromAzure(clusterLister ClusterLister) (*Map, error) {
	listResult, err := clusterLister.List()
	if err != nil {
		log.Printf("Parse Map From Azure: %v", err)
		return nil, err
	}

	return &Map{nameMap: clusterNamesToMap(listResult.Value)}, nil
}

// ClusterByName returns the cluster of the given cluster name. Returns nil and false if no
// cluster with the given name exists, non-nil and true otherwise
func (m Map) ClusterByName(name string) (*containerservice.ContainerService, bool) {
	cl, found := m.nameMap[name]
	return cl, found
}

// ClusterNamesByVersion returns a slice of all cluster names which match a given cluster version
func (m Map) ClusterNamesByVersion(matchingVersion string) []string {
	// var ret []string
	// for name, cluster := range m.nameMap {
	// 	cluster.
	// 	if matchingVersion == cluster.CurrentNodeVersion {
	// 		ret = append(ret, name)
	// 	}
	// }
	// return ret
	return nil
}

// Names returns all cluster names in the map. The order of the returned slice is undefined
func (m Map) Names() []string {
	ret := make([]string, len(m.nameMap))
	i := 0
	for name := range m.nameMap {
		ret[i] = name
		i++
	}
	return ret
}
