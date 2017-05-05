package azure

import (
	"github.com/Azure/azure-sdk-for-go/arm/containerservice"
)

// ClusterLister is an interface for listing Azure clusters. It has an adapter for the
// standard *(github.com/Azure/azure-sdk-for-go/arm/conatinerservice).ListResult as well as a fake implementation,
// to be used in unit tests. Use this as a parameter in your funcs so that they can be more
// easily unit tested
type ClusterLister interface {
	List() (*containerservice.ListResult, error)
}
