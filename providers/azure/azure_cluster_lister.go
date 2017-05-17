package azure

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/arm/containerservice"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/deis/k8s-claimer/config"
)

// AzureClusterLister is a ClusterLister implementation that uses the Azure Go SDK to list clusters
// on a live Azure cluster
type AzureClusterLister struct {
	Config *config.Azure
}

// NewAzureClusterLister creates a new AzureClusterLister configured to use the given client.
func NewAzureClusterLister(azureConfig *config.Azure) *AzureClusterLister {
	return &AzureClusterLister{Config: azureConfig}
}

// List is the ClusterLister interface implementation
func (a *AzureClusterLister) List() (*containerservice.ListResult, error) {
	bearerAuthorizer, err := NewBearerAuthorizer(a.Config, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Printf("Error trying to create Bearer Authorizer: %s", err)
		return nil, err
	}

	csClient := containerservice.NewContainerServicesClient(a.Config.SubscriptionID)
	csClient.Authorizer = bearerAuthorizer
	listResult, err := csClient.List()
	if err != nil {
		log.Printf("Error trying to fetch Azure Cluster List: %s\n", err)
		return nil, err
	}
	return &listResult, nil
}
