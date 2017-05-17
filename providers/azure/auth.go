package azure

import (
	"fmt"
	"log"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/deis/k8s-claimer/config"
)

const (
	credentialsPath = "/.azure/credentials.json"
)

// NewBearerAuthorizer creates a new BearerAuthorizer using values of the passed credentials map.
func NewBearerAuthorizer(a *config.Azure, scope string) (*autorest.BearerAuthorizer, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, a.TenantID)
	if err != nil {
		log.Printf("Error trying to create new OAuth Config:%s\n", err)
		return nil, err
	}
	token, err := adal.NewServicePrincipalToken(*oauthConfig, a.ClientID, a.ClientSecret, scope)
	if err != nil {
		log.Printf("Error trying to create New Service Principal Token:%s", err)
		return nil, err
	}
	return autorest.NewBearerAuthorizer(token), nil
}

func ensureValueStrings(mapOfInterface map[string]interface{}) map[string]string {
	mapOfStrings := make(map[string]string)
	for key, value := range mapOfInterface {
		mapOfStrings[key] = ensureValueString(value)
	}
	return mapOfStrings
}

func ensureValueString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
